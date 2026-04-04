#!/usr/bin/env node

/**
 * review-poll.js — Poll a GitHub PR for the Qodo code review comment.
 *
 * Qodo posts a placeholder comment immediately after PR creation, then
 * edits it with the full review once analysis is complete. This script
 * polls until the final review appears and prints it to stdout.
 *
 * Usage:
 *   node scripts/review-poll.js <pr-url> [--interval 15] [--timeout 300]
 *
 * Accepts:
 *   https://github.com/owner/repo/pull/123
 *   owner/repo#123
 *
 * Exit codes:
 *   0 — review found (body printed to stdout)
 *   1 — timeout
 *   2 — argument/usage error
 */

import { execFileSync } from "node:child_process";

const PLACEHOLDER_MARKERS = [
  "Looking for bugs?",
  "Check back in a few minutes",
];

const REVIEW_READY_MARKERS = [
  "Bugs",
  "Rule violations",
  "Action required",
  "Review recommended",
  "Requirement gaps",
  "UX Issues",
];

function usage() {
  console.error(
    `Usage: node scripts/review-poll.js <pr-url> [--interval 15] [--timeout 300]

Arguments:
  pr-url      GitHub PR URL or owner/repo#number

Options:
  --interval  Seconds between polls (default: 15)
  --timeout   Max seconds to wait (default: 300)`
  );
  process.exit(2);
}

function parsePR(input) {
  // https://github.com/owner/repo/pull/123
  const urlMatch = input.match(
    /github\.com\/([^/]+)\/([^/]+)\/pull\/(\d+)/
  );
  if (urlMatch) {
    return { owner: urlMatch[1], repo: urlMatch[2], number: urlMatch[3] };
  }

  // owner/repo#123
  const shortMatch = input.match(/^([^/]+)\/([^#]+)#(\d+)$/);
  if (shortMatch) {
    return { owner: shortMatch[1], repo: shortMatch[2], number: shortMatch[3] };
  }

  return null;
}

const VALID_SLUG = /^[A-Za-z0-9_.-]+$/;

function validatePR({ owner, repo, number }) {
  if (!VALID_SLUG.test(owner) || !VALID_SLUG.test(repo)) {
    console.error(`Error: invalid owner/repo characters: ${owner}/${repo}`);
    process.exit(2);
  }
  if (!/^\d+$/.test(number)) {
    console.error(`Error: invalid PR number: ${number}`);
    process.exit(2);
  }
}

function parseArgs(argv) {
  const args = argv.slice(2);
  if (args.length === 0 || args.includes("--help") || args.includes("-h")) {
    usage();
  }

  let prInput = null;
  let interval = 15;
  let timeout = 300;

  for (let i = 0; i < args.length; i++) {
    if (args[i] === "--interval" && args[i + 1]) {
      interval = parseInt(args[++i], 10);
    } else if (args[i] === "--timeout" && args[i + 1]) {
      timeout = parseInt(args[++i], 10);
    } else if (!args[i].startsWith("--")) {
      prInput = args[i];
    }
  }

  if (!prInput) usage();

  const pr = parsePR(prInput);
  if (!pr) {
    console.error(`Error: could not parse PR reference: ${prInput}`);
    process.exit(2);
  }
  validatePR(pr);

  if (!Number.isFinite(interval) || interval < 1) {
    console.error(`Error: --interval must be a positive integer (got: ${interval})`);
    process.exit(2);
  }
  if (!Number.isFinite(timeout) || timeout < 1) {
    console.error(`Error: --timeout must be a positive integer (got: ${timeout})`);
    process.exit(2);
  }

  return { ...pr, interval, timeout };
}

function fetchComments(owner, repo, number) {
  const result = execFileSync(
    "gh",
    [
      "api",
      `repos/${owner}/${repo}/issues/${number}/comments`,
      "--paginate",
      "--jq",
      ".[] | {id, author: .user.login, body}",
    ],
    { encoding: "utf-8", stdio: ["pipe", "pipe", "pipe"] }
  );
  // --paginate with --jq outputs one JSON object per line (NDJSON)
  const lines = result.trim().split("\n").filter(Boolean);
  return lines.map((line) => JSON.parse(line));
}

function isPlaceholder(body) {
  return PLACEHOLDER_MARKERS.some((m) => body.includes(m));
}

function isReviewReady(body) {
  return (
    !isPlaceholder(body) &&
    REVIEW_READY_MARKERS.some((m) => body.includes(m))
  );
}

function findQodoReviewComment(comments) {
  // Qodo posts two comments: "Review Summary" and "Code Review".
  // We want the "Code Review" one (contains bug/rule findings).
  // Look for the Qodo comment that is either:
  //   - the code review (contains "Code Review" in body)
  //   - the placeholder (contains "Looking for bugs?")
  const qodoComments = comments.filter(
    (c) => c.author && c.author.toLowerCase().includes("qodo")
  );

  // Prefer the "Code Review" comment
  const codeReview = qodoComments.find((c) => c.body.includes("Code Review"));
  if (codeReview) return codeReview;

  // Fall back to any Qodo comment that looks like a placeholder
  const placeholder = qodoComments.find((c) => isPlaceholder(c.body));
  if (placeholder) return placeholder;

  // Fall back to the last Qodo comment (most likely the review)
  return qodoComments.length > 0 ? qodoComments[qodoComments.length - 1] : null;
}

function sleep(seconds) {
  return new Promise((resolve) => setTimeout(resolve, seconds * 1000));
}

async function main() {
  const { owner, repo, number, interval, timeout } = parseArgs(process.argv);
  const prRef = `${owner}/${repo}#${number}`;

  console.error(`Polling ${prRef} every ${interval}s (timeout: ${timeout}s)...`);

  const start = Date.now();

  while ((Date.now() - start) / 1000 < timeout) {
    try {
      const comments = fetchComments(owner, repo, number);
      const qodo = findQodoReviewComment(comments);

      if (qodo) {
        if (isReviewReady(qodo.body)) {
          console.error(`Review ready for ${prRef}`);
          console.log(qodo.body);
          process.exit(0);
        }
        console.error(
          `Found Qodo comment but still placeholder... (${Math.round((Date.now() - start) / 1000)}s elapsed)`
        );
      } else {
        console.error(
          `No Qodo comment yet... (${Math.round((Date.now() - start) / 1000)}s elapsed)`
        );
      }
    } catch (err) {
      console.error(`API error: ${err.message}, retrying...`);
    }

    await sleep(interval);
  }

  console.error(`Timeout: no Qodo review found after ${timeout}s`);
  process.exit(1);
}

main();
