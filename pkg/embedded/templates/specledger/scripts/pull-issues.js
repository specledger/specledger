#!/usr/bin/env node

/**
 * pull-issues.js
 * Pull beads issues from Supabase and rebuild .beads/issues.jsonl
 *
 * Usage:
 *   node scripts/pull-issues.js --repo-owner <owner> --repo-name <name>
 *
 * Features:
 *   - Fetches all issues from Supabase bd_issues table
 *   - Rebuilds .beads/issues.jsonl in correct JSONL format
 *   - Preserves dependencies and comments
 *   - Beads daemon auto-detects changes and reimports to SQLite
 */

import { createClient } from '@supabase/supabase-js'
import fs from 'fs'
import path from 'path'
import os from 'os'

const BEADS_FILE = path.join(process.cwd(), '.beads', 'issues.jsonl')
const CREDENTIALS_FILE = path.join(os.homedir(), '.specledger', 'credentials.json')
const SUPABASE_URL = 'https://lmjpnzplurfnojfqtqly.supabase.co'

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
  console.log('🔄 Pulling beads issues from Supabase...\n')

  const accessToken = loadAccessToken()
  const cmdArgs = parseArgs()

  const repoOwner = cmdArgs['repo-owner']
  const repoName = cmdArgs['repo-name']

  if (!repoOwner || !repoName) {
    console.error('❌ Missing repository information')
    console.error('Usage: node scripts/pull-issues.js --repo-owner <owner> --repo-name <name>')
    process.exit(1)
  }

  const supabase = createClient(SUPABASE_URL, accessToken)

  const { data: project, error: projectError } = await supabase
    .from('projects')
    .select('id')
    .eq('repo_owner', repoOwner)
    .eq('repo_name', repoName)
    .single()

  if (projectError || !project) {
    console.error('❌ Project not found:', repoOwner + '/' + repoName)
    console.error(projectError)
    process.exit(1)
  }

  console.log(`✓ Found project: ${repoOwner}/${repoName} (${project.id})`)

  // Fetch all issues
  const { data: issues, error: issuesError } = await supabase
    .from('bd_issues')
    .select('*')
    .eq('project_id', project.id)
    .order('created_at', { ascending: true })

  if (issuesError) {
    console.error('❌ Failed to fetch issues:', issuesError)
    process.exit(1)
  }

  console.log(`✓ Fetched ${issues.length} issues`)

  // Fetch all dependencies
  const { data: dependencies, error: depsError } = await supabase
    .from('bd_dependencies')
    .select('*')
    .eq('project_id', project.id)

  if (depsError) {
    console.warn('⚠️  Failed to fetch dependencies:', depsError)
  }

  console.log(`✓ Fetched ${dependencies?.length || 0} dependencies`)

  // Fetch all comments
  const { data: comments, error: commentsError } = await supabase
    .from('bd_comments')
    .select('*')
    .eq('project_id', project.id)

  if (commentsError) {
    console.warn('⚠️  Failed to fetch comments:', commentsError)
  }

  console.log(`✓ Fetched ${comments?.length || 0} comments\n`)

  // Group dependencies by issue_id
  const depsByIssue = {}
  if (dependencies) {
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
  }

  // Group comments by issue_id
  const commentsByIssue = {}
  if (comments) {
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
  }

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
  console.log(`   - Total dependencies: ${dependencies?.length || 0}`)
  console.log(`   - Total comments: ${comments?.length || 0}`)

  console.log('\n✅ Done! Beads daemon will auto-import changes.')
  console.log('   Run: bd ready')
}

main().catch(err => {
  console.error('❌ Error:', err.message)
  process.exit(1)
})
