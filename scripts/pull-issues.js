#!/usr/bin/env node

/**
 * pull-issues.js
 * Pull issues from Supabase and rebuild .beads/issues.jsonl
 *
 * Usage:
 *   node scripts/pull-issues.js --repo-owner <owner> --repo-name <name>
 *
 * Features:
 *   - Uses CURL (no npm dependencies required)
 *   - Fetches all issues from Supabase issues table
 *   - Rebuilds .beads/issues.jsonl in correct JSONL format
 *   - Preserves dependencies and comments
 *   - Beads daemon auto-detects changes and reimports to SQLite
 */

import { execSync } from 'child_process'
import fs from 'fs'
import path from 'path'
import os from 'os'

const BEADS_FILE = path.join(process.cwd(), '.beads', 'issues.jsonl')
const CREDENTIALS_FILE = path.join(os.homedir(), '.specledger', 'credentials.json')

// Get Supabase config - uses same env vars as sl CLI (see pkg/cli/auth/client.go)
function getSupabaseConfig() {
  const url = process.env.SPECLEDGER_SUPABASE_URL || 'https://iituikpbiesgofuraclk.supabase.co'
  const key = process.env.SPECLEDGER_SUPABASE_ANON_KEY || 'sb_publishable_KpaZ2lKPu6eJ5WLqheu9_A_J9dYhGQb'
  return { url, key }
}

// Load access token from credentials file
function loadAccessToken() {
  if (!fs.existsSync(CREDENTIALS_FILE)) {
    console.error('❌ Credentials file not found:', CREDENTIALS_FILE)
    console.error('   Run "sl login" to authenticate first.')
    process.exit(1)
  }

  try {
    const credentials = JSON.parse(fs.readFileSync(CREDENTIALS_FILE, 'utf8'))
    if (!credentials.access_token) {
      console.error('❌ No access_token found in credentials file')
      console.error('   Run "sl login" to authenticate first.')
      process.exit(1)
    }
    return credentials.access_token
  } catch (err) {
    console.error('❌ Failed to read credentials file:', err.message)
    process.exit(1)
  }
}

// Execute CURL request to Supabase REST API
function supabaseQuery(endpoint, params = {}) {
  const { url, key } = getSupabaseConfig()
  const accessToken = loadAccessToken()

  // Build query string
  const queryParts = Object.entries(params).map(([k, v]) => `${k}=${encodeURIComponent(v)}`)
  const queryString = queryParts.length > 0 ? '?' + queryParts.join('&') : ''
  const fullUrl = `${url}/rest/v1/${endpoint}${queryString}`

  try {
    const result = execSync(`curl -s "${fullUrl}" -H "apikey: ${key}" -H "Authorization: Bearer ${accessToken}"`, {
      encoding: 'utf8',
      maxBuffer: 50 * 1024 * 1024 // 50MB buffer for large responses
    })
    return JSON.parse(result)
  } catch (err) {
    console.error(`❌ CURL request failed: ${err.message}`)
    return null
  }
}

// Parse command line arguments
function parseArgs() {
  const args = process.argv.slice(2)
  const params = {}

  for (let i = 0; i < args.length; i += 2) {
    const key = args[i].replace(/^--/, '')
    const value = args[i + 1]
    params[key] = value
  }

  return params
}

async function main() {
  console.log('🔄 Pulling issues from Supabase...\n')

  const { url } = getSupabaseConfig()
  const cmdArgs = parseArgs()

  const repoOwner = cmdArgs['repo-owner']
  const repoName = cmdArgs['repo-name']

  if (!repoOwner || !repoName) {
    console.error('❌ Missing repository information')
    console.error('Usage: node scripts/pull-issues.js --repo-owner <owner> --repo-name <name>')
    process.exit(1)
  }

  console.log(`🔗 Using Supabase: ${url}`)

  // Step 1: Find project
  const projects = supabaseQuery('projects', {
    select: 'id',
    repo_owner: `eq.${repoOwner}`,
    repo_name: `eq.${repoName}`
  })

  if (!projects || projects.length === 0) {
    console.error('❌ Project not found:', repoOwner + '/' + repoName)
    process.exit(1)
  }

  const project = projects[0]
  console.log(`✓ Found project: ${repoOwner}/${repoName} (${project.id})`)

  // Step 2: Find all specs for this project
  const specs = supabaseQuery('specs', {
    select: 'id,spec_key',
    project_id: `eq.${project.id}`
  })

  if (!specs || specs.length === 0) {
    console.log('ℹ️  No specs found for this project')
    process.exit(0)
  }

  const specIds = specs.map(s => s.id)
  console.log(`✓ Found ${specs.length} specs`)

  // Step 3: Fetch all issues for these specs
  const issues = supabaseQuery('issues', {
    select: '*',
    spec_id: `in.(${specIds.join(',')})`,
    order: 'created_at.asc'
  })

  if (!issues) {
    console.error('❌ Failed to fetch issues')
    process.exit(1)
  }

  console.log(`✓ Fetched ${issues.length} issues`)

  const issueIds = issues.map(i => i.id)

  // Step 4: Fetch dependencies
  let dependencies = []
  if (issueIds.length > 0) {
    const deps = supabaseQuery('dependencies', {
      select: '*',
      issue_id: `in.(${issueIds.join(',')})`
    })
    if (deps) {
      dependencies = deps
    }
  }
  console.log(`✓ Fetched ${dependencies.length} dependencies`)

  // Step 5: Fetch comments
  let comments = []
  if (issueIds.length > 0) {
    const cmts = supabaseQuery('comments', {
      select: '*',
      issue_id: `in.(${issueIds.join(',')})`
    })
    if (cmts) {
      comments = cmts
    }
  }
  console.log(`✓ Fetched ${comments.length} comments\n`)

  // Group dependencies by issue_id
  const depsByIssue = {}
  dependencies.forEach(dep => {
    if (!depsByIssue[dep.issue_id]) {
      depsByIssue[dep.issue_id] = []
    }
    depsByIssue[dep.issue_id].push({
      issue_id: dep.issue_id,
      depends_on_id: dep.depends_on_id,
      type: dep.type,
      created_at: dep.created_at,
      created_by: dep.created_by || 'unknown'
    })
  })

  // Group comments by issue_id
  const commentsByIssue = {}
  comments.forEach(comment => {
    if (!commentsByIssue[comment.issue_id]) {
      commentsByIssue[comment.issue_id] = []
    }
    commentsByIssue[comment.issue_id].push({
      id: comment.id,
      issue_id: comment.issue_id,
      author: comment.author,
      text: comment.text,
      created_at: comment.created_at
    })
  })

  // Build JSONL lines
  const lines = issues.map(issue => {
    const obj = {
      id: issue.id,
      title: issue.title,
      status: issue.status,
      priority: issue.priority,
      issue_type: issue.issue_type,
      created_at: issue.created_at,
      updated_at: issue.updated_at
    }

    // Add optional fields only if they exist
    if (issue.description) obj.description = issue.description
    if (issue.design) obj.design = issue.design
    if (issue.acceptance_criteria) obj.acceptance_criteria = issue.acceptance_criteria
    if (issue.closed_at) obj.closed_at = issue.closed_at
    if (issue.labels && issue.labels.length > 0) obj.labels = issue.labels

    // Add dependencies if any
    if (depsByIssue[issue.id] && depsByIssue[issue.id].length > 0) {
      obj.dependencies = depsByIssue[issue.id]
    }

    // Add comments if any
    if (commentsByIssue[issue.id] && commentsByIssue[issue.id].length > 0) {
      obj.comments = commentsByIssue[issue.id]
    }

    return JSON.stringify(obj)
  })

  // Write to .beads/issues.jsonl
  const content = lines.join('\n') + '\n'

  // Backup existing file
  if (fs.existsSync(BEADS_FILE)) {
    const backupFile = BEADS_FILE + '.backup-' + Date.now()
    fs.copyFileSync(BEADS_FILE, backupFile)
    console.log(`📦 Backed up existing file to ${path.basename(backupFile)}`)
  }

  fs.writeFileSync(BEADS_FILE, content, 'utf8')
  console.log(`✓ Wrote ${issues.length} issues to .beads/issues.jsonl\n`)

  // Stats
  const withDeps = Object.keys(depsByIssue).length
  const withComments = Object.keys(commentsByIssue).length
  console.log('📊 Summary:')
  console.log(`   - Issues: ${issues.length}`)
  console.log(`   - With dependencies: ${withDeps}`)
  console.log(`   - With comments: ${withComments}`)
  console.log(`   - Total dependencies: ${dependencies.length}`)
  console.log(`   - Total comments: ${comments.length}`)

  console.log('\n✅ Done! Beads daemon will auto-import changes.')
  console.log('   Run: bd ready')
}

main().catch(err => {
  console.error('❌ Error:', err.message)
  process.exit(1)
})
