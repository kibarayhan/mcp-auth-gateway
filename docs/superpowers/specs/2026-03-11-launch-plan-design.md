# Design: MCP Auth Gateway Launch Plan

**Date:** 2026-03-11
**Status:** Draft
**Type:** Go-to-market / monetization plan
**Timeline:** 12 weeks
**Approach:** Community-first with security audit content funnel

## Problem

The MCP Auth Gateway MVP is complete (all 6 modules shipped) but the repo is private with zero visibility. We need to validate demand, identify customers, and reach $100+ MRR within 12 weeks.

## Strategy: Security Audit Funnel

Systematically audit popular MCP servers for security gaps, publish findings as video + written content, and position the gateway as the solution. Each audit serves as both marketing and customer research — people who engage self-identify as the target market.

## Section 1: Customer Discovery Engine

### Audit Pipeline
1. Identify the top 50 most-used MCP servers (by GitHub stars, npm downloads, community mentions)
2. Run each through the gateway with PII filter + audit logging enabled
3. Document what leaks: PII in responses, missing auth, no rate limiting, dangerous tool arguments
4. Package findings as short-form video (3-5 min) + written post

### Discovery Mechanism
- People who watch/share/comment on audits self-identify as the target market
- Track engagement: who replies, what questions they ask, what MCP servers they use
- Every DM asking "how do I set this up for my team?" is a potential paying customer

### Volume Target
- 2 audits per week for the first 8 weeks = 16 audits covering major MCP servers
- Each audit names the specific risks and shows the gateway solving them live

### Channels
- YouTube for video
- X/Twitter threads summarizing findings
- Reddit (r/ClaudeAI, r/LocalLLaMA, r/devops)
- Hacker News for the splashiest findings

## Section 2: Open Source Launch Strategy

### Pre-launch Prep (Week 1)
- Push repo to GitHub public
- Clean up README with clear "what this does" + 30-second GIF showing it in action
- Add a quickstart.md — get the gateway running in under 2 minutes with a single MCP server
- License is already MIT

### Launch Sequence (Week 2)
- First security audit video drops the same day the repo goes public
- Post to Hacker News: frame as "I audited 10 popular MCP servers for security — here's what I found" (not "here's my product")
- Cross-post to r/ClaudeAI, r/devops, X
- The gateway is the tool used in the audit, not the pitch

### Building in Public Cadence (Weeks 2-8)
- 2x/week: short X posts showing progress, findings, user questions
- 1x/week: security audit video
- Engage every GitHub issue and star — early community members become advocates

### Key Principle
Lead with the audits, not the product. People share "MCP servers leak your email addresses" — they don't share "here's a proxy tool."

## Section 3: Monetization Path

### Free Tier (open source, always)
- All 6 modules, unlimited servers
- File-based audit log
- API key auth
- Already built — no extra work

### Pro Tier ($49/month) — build only when 5+ people ask
- OAuth 2.0 / OIDC integration (Okta, Auth0, Azure AD)
- S3/webhook audit log export
- Email alerts on policy violations
- Priority GitHub issues

### Sales Process (Weeks 8-12)
- By week 8, have a list of engaged users from audit content + GitHub issues
- Direct outreach: "Hey, I noticed you're using the gateway with [X]. Would [specific feature] help your team?"
- Offer 2-week free trial of Pro features
- Use Stripe or Lemon Squeezy for billing
- Target: 2-3 paying users by week 12 = $100-150 MRR

### NOT Building Yet
- No web dashboard
- No hosted/managed version
- No Enterprise tier
- No SAML/mTLS
- Only build what people actually ask and pay for

## Section 4: Metrics & Decision Points

### Weekly Tracking
- GitHub stars + new issues/PRs
- Video views + engagement (comments, shares)
- DMs / emails from potential users
- Audit pipeline: servers analyzed vs. servers remaining

### Decision Gates

| Week | Check | If YES | If NO |
|------|-------|--------|-------|
| 4 | 50+ stars, 3+ inbound conversations? | Continue, double down on best content | Shift audit angle — try different MCP categories or channels |
| 8 | 200+ stars, 5+ people asking for features? | Start building Pro tier | Pause Pro, do 5 direct user interviews to understand why not |
| 12 | 2+ paying customers, $100+ MRR? | Plan Enterprise tier, consider hosted version | Re-evaluate pricing, target market, or pivot approach |

### Kill Criteria
If by week 8: <50 stars and zero inbound interest, the market isn't ready or the positioning is wrong. Doesn't mean the product is bad — might mean MCP adoption needs more time.

## Section 5: Budget (~$28/month)

| Item | Cost | Purpose |
|------|------|---------|
| NotebookLM | Free | Audio narrative for audit videos |
| OBS Studio | Free | Screen recording |
| CapCut / DaVinci Resolve | Free | Combine audio + screen recordings |
| Domain + landing page | ~$20/month | Credibility, email capture |
| Lemon Squeezy / Stripe | Free until revenue | Payment processing |
| X/Twitter Blue | $8/month | Longer posts, better reach |

### Video Production Pipeline
1. NotebookLM generates audio narrative from audit notes
2. OBS captures terminal screen recordings (gateway in action, PII redaction, logs)
3. Free video editor combines audio + screen recordings
4. No manual narration needed — speeds up production significantly

## Week-by-Week Summary

| Week | Focus | Key Deliverable |
|------|-------|-----------------|
| 1 | Prep | Repo public, README polished, quickstart guide |
| 2 | Launch | First audit video + HN post + Reddit |
| 3-4 | Content | 4 more audit videos, engage community |
| 5-8 | Growth | Continue audits, build in public, track who engages |
| 8 | Gate | Decide: build Pro tier or interview users |
| 9-11 | Monetize | Build Pro features (if validated), direct outreach |
| 12 | Gate | Decide: scale up or pivot |
