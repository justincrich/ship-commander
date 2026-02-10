# Ship Commander TUI: User Personas & Mental Models

**Version**: 1.0
**Date**: 2025-02-09
**Designer**: TUI Designer Agent
**Status**: Draft for Product Manager Review

---

## Executive Summary

Ship Commander serves **technical users** who are comfortable with CLI tools but want the *power of AI agents* without losing *control and predictability*. The TUI must balance two opposing needs:

1. **Starfleet Admiral Experience**: Feeling of command, oversight, strategic decision-making
2. **Developer Control**: Transparency, debuggability, manual intervention when needed

This document defines three primary personas that span the user base, maps their mental models, and identifies emotional touchpoints in the user journey.

---

## Table of Contents

1. [User Personas](#user-personas)
2. [Mental Models & Concept Mapping](#mental-models--concept-mapping)
3. [User Journeys](#user-journeys)
4. [Emotional Touchpoints](#emotional-touchpoints)
5. [Persona-Specific UI Needs](#persona-specific-ui-needs)

---

## User Personas

### Persona 1: The Staff Commander (Power User)

**WHO**: Senior developer or tech lead, comfortable with CLI tools, manages multiple features

**Demographics**:
- **Role**: Staff Engineer, Tech Lead, Senior Developer
- **Experience**: 7-15 years in software development
- **Terminal proficiency**: High (vim, tmux, complex CLI workflows daily)
- **Star Trek familiarity**: Medium (knows the show, appreciates the aesthetic)
- **AI tooling experience**: Early adopter (uses Cursor, Copilot, Claude Code regularly)

**Goals**:
1. **Orchestrate complex features** without losing visibility into implementation
2. **Maintain control** over what AI agents produce (can inspect, halt, retry)
3. **Parallelize work** across multiple directives efficiently
4. **Debug agent failures** quickly when things go wrong
5. **Ship faster** through automation while maintaining quality standards

**Pain Points**:
1. **Black-box AI tools** that don't show what they're doing
2. **Context switching** between terminal, browser, and editor to monitor agents
3. **Manual merge conflicts** when parallel work collides
4. **No visibility** into which agent is stuck and why
5. **Hard to recover** from agent failures (manual cleanup)

**What "Starfleet Admiral" means to them**:
- **Command deck**: Feeling of overseeing multiple "away teams" (agents)
- **Strategic oversight**: Seeing the "big picture" of waves, dependencies, health
- **Trusted crew**: Confidence that agents (crew) follow protocols (gates)
- **Red alert**: Clear signaling when something needs human intervention

**Mental Model**:
```
Directives = Mission orders (tasks to complete)
Agents = Ensigns/Crew (execute the orders)
Gates = Protocol checks (quality assurance before proceeding)
Waves = Formation groups (batches of missions executed in parallel)
Doctor = Chief Medical Officer (monitors crew health, detects issues)
Propulsion = Warp drive (pulls missions and dispatches crew)
Captain = Me (operator making strategic decisions)
```

**Expected TUI Features**:
- **Real-time agent monitoring**: Which agents are working, on what, elapsed time
- **Phase tracking**: RED→GREEN→REFACTOR progress per directive
- **Gate results**: Immediate visibility into test failures, lint errors, type errors
- **Intervention controls**: Halt, retry, approve commands at keyboard speed
- **Dependency visualization**: See which directives are blocked and why
- **Health dashboard**: System status (stuck agents, failed gates, merge conflicts)
- **Log streaming**: Scrollable event log with severity filtering

---

### Persona 2: The Junior Lieutenant (Novice User)

**WHO**: Junior to mid-level developer, learning the system, needs guidance

**Demographics**:
- **Role**: Junior Developer, Mid-level Developer
- **Experience**: 1-5 years in software development
- **Terminal proficiency**: Medium (basic commands, learning advanced workflows)
- **Star Trek familiarity**: Low to Medium (seen a few episodes, gets the references)
- **AI tooling experience**: Casual user (Copilot for autocomplete, exploring Claude Code)

**Goals**:
1. **Learn the system** without breaking production
2. **Understand what's happening** at each step (no mystery)
3. **Get help** when stuck (clear error messages, contextual help)
4. **Contribute features** without deep knowledge of the architecture
5. **Build confidence** through successful deployments

**Pain Points**:
1. **Overwhelming dashboards** with too much information at once
2. **Unclear terminology** (what's a "directive"? what's a "gate"?)
3. **Fear of breaking things** (accidentally halting production work)
4. **No onboarding** (learning by trial-and-error)
5. **Imposter syndrome** (everyone else seems to know what they're doing)

**What "Starfleet Admiral" means to them**:
- **Academy training**: Learning the ropes in a safe environment
- **Supportive crew**: Helpful feedback, clear instructions
- **Graduated responsibility**: Start small (single directives), work up to waves
- **No shame in asking**: Help text, tooltips, contextual guidance

**Mental Model** (simplified):
```
Directives = Tasks on my todo list
Agents = AI helpers that write code for me
Gates = Tests that must pass before I can commit
Waves = Groups of related tasks
Doctor = System that checks if helpers are stuck
Propulsion = System that starts helpers on tasks
Captain = The senior dev who helps me when I'm stuck
```

**Expected TUI Features**:
- **Onboarding overlay**: First-run tour explaining panels and concepts
- **Simplified views**: Toggle between "basic" and "advanced" modes
- **Inline help**: Contextual hints (press `?` for help, `h` to halt)
- **Safe defaults**: Can't accidentally halt production without confirmation
- **Progress indicators**: Clear "what's happening now" displays
- **Error explanations**: Human-readable messages, not stack traces
- **Undo/redo**: Can recover from mistakes (retry, revert)

---

### Persona 3: The Fleet Admiral (Busy Manager/Architect)

**WHO**: Engineering manager, architect, or principal overseeing multiple projects

**Demographics**:
- **Role**: Engineering Manager, Software Architect, Principal Engineer
- **Experience**: 15+ years in software development
- **Terminal proficiency**: Low to Medium (uses CLI occasionally, prefers higher-level views)
- **Star Trek familiarity**: Low (in it for the productivity, not the theme)
- **AI tooling experience**: Strategic user (evaluates tools for team adoption)

**Goals**:
1. **Track progress** across multiple projects/features without micromanaging
2. **Identify blockers** quickly (which directives are stuck?)
3. **Ensure quality** (are gates passing? is coverage adequate?)
4. **Plan sprints** based on velocity and wave completion
5. **Onboard team members** efficiently (consistent workflows)

**Pain Points**:
1. **Too much detail** in existing dashboards (don't need to see every gate result)
2. **No aggregated view** (how is the overall project doing?)
3. **Hard to spot patterns** (are directives failing for the same reason?)
4. **No forecasting** (when will wave 2 complete based on current velocity?)
5. **Team velocity opaque** (how many directives completed this week?)

**What "Starfleet Admiral" means to them**:
- **Fleet status**: High-level view of all "ships" (projects)
- **Strategic decisions**: Which features to prioritize, which to defer
- **Resource allocation**: How many agents (crew) are available vs. needed
- **Mission reports**: Summarized outcomes, not granular logs

**Mental Model** (abstracted):
```
Directives = Feature tasks (don't care about implementation details)
Agents = Compute resources (how many are running? what's utilization?)
Gates = Quality gates (pass/fail, don't need to see every error)
Waves = Sprint batches (what's in this sprint vs. next?)
Doctor = Health monitoring (is the system degraded?)
Propulsion = Execution engine (is it working?)
Captain = Team leads (are they making good decisions?)
```

**Expected TUI Features**:
- **Executive dashboard**: High-level metrics (velocity, success rate, blocked tasks)
- **Aggregated views**: Roll up by project, wave, feature area
- **Trend visualization**: Velocity over time, common failure modes
- **Alerting**: Notifications for critical issues (stuck agents, merge conflicts)
- **Drill-down**: Click into details only when needed
- **Multi-project view**: See all projects in one TUI
- **Reporting**: Export summaries for standups/status reports

---

## Mental Models & Concept Mapping

### Core Concept Translation

All personas share a need to map **Ship Commander terminology** to their **existing mental models**. The TUI should reinforce these mappings through consistent language and visual metaphors.

| Ship Commander Term | Staff Commander Mental Model | Junior Lieutenant Mental Model | Fleet Admiral Mental Model |
|---------------------|------------------------------|--------------------------------|----------------------------|
| **Directive** | Mission order, work item | Task, TODO item | Feature task, story |
| **Agent** | Ensign, away team member | AI helper, bot | Compute resource, worker |
| **Gate** | Protocol check, validation | Test, quality check | Quality gate, milestone |
| **Wave** | Formation group, execution batch | Batch of tasks | Sprint, iteration |
| **Doctor** | Chief Medical Officer | Health checker, monitor | Health monitoring system |
| **Propulsion** | Warp drive, mission control | Task starter, dispatcher | Execution engine |
| **Captain** | Me (the operator) | The senior dev | Team lead |
| **Intake** | Mission briefing, planning | Setup, preparation | Planning phase |
| **Synthesis** | Plan approval, coordination | Review, confirmation | Approval gate |
| **Review** | Pre-merge inspection, QA | Code review | Approval checkpoint |
| **Approved** | Ready for merge, ship it | Can merge | Approved for integration |
| **Halted** | Red alert, emergency stop | Stuck, broken | Blocked, failed |
| **Stuck** | Disabled, needs attention | Waiting for help | Degraded, intervention needed |
| **Worktree** | Away team staging area | Workspace | Isolated environment |
| **Merge** | Return to base, report back | Combine work | Integration, deployment |

### Visual Metaphors by Persona

#### Staff Commander (Power User)
- **Mission Control**: Literal "bridge" with stations for each subsystem
- **Color coding**: LCARS orange/blue/purple for instant status recognition
- **Animations**: Warp engage (dispatch), transporter (merge), shields (gates)
- **Sound effects**: Optional beep on red alert, chime on success
- **Terminology**: Full Star Trek (engage, red alert, shields holding)

#### Junior Lieutenant (Novice User)
- **Mission Control**: Simplified dashboard with clear labels
- **Color coding**: Green (good), yellow (warning), red (error) - less LCARS
- **Animations**: Subtle (don't overwhelm), can be disabled
- **Sound effects**: Off by default, can enable
- **Terminology**: Hybrid ("Task #42 is running" vs. "Directive #42 is executing")

#### Fleet Admiral (Manager)
- **Mission Control**: Executive summary view, minimal decoration
- **Color coding**: Status-focused (red/green) - skip LCARS purple/pink
- **Animations**: Disabled by default (wastes time)
- **Sound effects**: Disabled
- **Terminology**: Business language ("Feature X is 80% complete")

### Mental Model Misalignments

Watch for these **conceptual gaps** where users might misinterpret the system:

1. **Directive vs. Branch**: Users may think "directive" = "git branch"
   - **Reality**: Directives create worktree branches, but directives are *higher-level* concepts
   - **TUI Fix**: Show branch name in directive card: "feature/42-auth-flow"

2. **Agent vs. Process**: Users may think "agent" = "background process"
   - **Reality**: Agents are *harness-managed* AI sessions, not system processes
   - **TUI Fix**: Show harness/model in agent card: "claude (opus)"

3. **Gate vs. Test**: Users may think "gate" = "just test suite"
   - **Reality**: Gates are *multi-stage* verification (typecheck, lint, build, tests)
   - **TUI Fix**: Show all gate stages in Phase Tracker: "d1 VERIFY_GREEN [typecheck✓ lint✓ tests⋯]"

4. **Wave vs. Sprint**: Users may think "wave" = "timeboxed sprint"
   - **Reality**: Waves are *dependency-based* execution groups, not timeboxed
   - **TUI Fix**: Show dependency relationships: "Wave 2: d3 (depends on d1,d2)"

5. **Doctor vs. Debugger**: Users may think "doctor" = "code debugger"
   - **Reality**: Doctor is *health monitoring* for agents, not a step-through debugger
   - **TUI Fix**: Clarify in log messages: "[WARN] doctor.stuck agent=cmdr-abc (health check, not code debug)"

---

## User Journeys

### Journey 1: Staff Commander - Complex Feature Intake

**User**: Staff Commander (power user)
**Goal**: Ingest a PRD for a multi-feature project (auth system + UI updates)
**Expected Outcome**: 12-15 directives created, dependency graph wired, ready for propulsion

#### Emotional Arc

| Phase | User Action | TUI Response | Emotional State | Design Intent |
|-------|-------------|--------------|-----------------|---------------|
| **Initiation** | Runs `ship-commander intake --prd ./auth-prd.md --with-tui` | Mission Briefing Room appears, LCARS orange border | "Let's do this" | **Excitement** - TUI launch with theme music (optional) |
| **Parsing** | Waits for PRD parse | Scanning animation [████████░░] 80% "Extracting use cases..." | "Hope this works" | **Anticipation** - Progress bar, phase label |
| **Analysis Start** | Sees specialist panels appear | Captain (DONE), Commander (WORKING...), Design Officer (WORKING...) | "Good, Captain validated the PRD" | **Confidence** - All specialists present |
| **Parallel Execution** | Watches real-time output | Commander: "Identifying infra tasks (12 found)...", Design Officer: "Scanning UI requirements..." | "Love seeing them work in parallel" | **Engagement** - 3 parallel panels updating live |
| **Synthesis** | Sees Synthesizer (WAITING) | Commander ✓ DONE, Design Officer ✓ DONE, Synthesizer "Merging outputs..." | "Almost there" | **Satisfaction** - Specialists completing |
| **Plan Review** | Sees proposed directives | "12 directives in 3 waves, 15/15 UCs covered ✓" | "That looks right" | **Validation** - Coverage summary is clear |
| **Dependency Graph** | Reviews tree view | ASCII tree shows Wave 1 (d1,d2), Wave 2 (d3→d1,d2) | "Dependencies look correct" | **Clarity** - Tree visualization makes sense |
| **Approval** | Presses `[a]` to approve | Transporter effect: "COORDINATES LOCKED... RE-MATERIALIZING... ARRIVED ✓" | "🚀 Let's go" | **Celebration** - Transporter animation delights |
| **Commit** | Sees confirmation | "✓ Created 12 directives in 3 waves" "✓ Wires 8 dependency edges" | "Ship it!" | **Accomplishment** - Clear success message |

#### Key Emotional Touchpoints

1. **Launch**: "Mission Briefing Room" theme + LCARS border → **"I'm on the bridge"**
2. **Parallel Execution**: Three panels updating live → **"My crew is working"**
3. **Coverage Validation**: "15/15 UCs covered ✓" → **"Nothing was missed"**
4. **Transporter Animation**: Dematerialize→rematerialize → **"Magic moment"**
5. **Success Message**: "12 directives created, 8 edges wired" → **"Ready to ship"**

#### Pain Points to Avoid

- **Long delays** without feedback (show "WORKING..." spinner after 10s)
- **Unclear progress** (always show phase + % complete)
- **Ambiguous errors** (show "Error: PRD has no use cases" not "Parse failed")
- **Hidden complexity** (show all specialist panels, not just one)

---

### Journey 2: Junior Lieutenant - First Directive Monitoring

**User**: Junior Lieutenant (novice)
**Goal**: Monitor their first AI-implemented directive without breaking it
**Expected Outcome**: Directive completes successfully, they learn the system

#### Emotional Arc

| Phase | User Action | TUI Response | Emotional State | Design Intent |
|-------|-------------|--------------|-----------------|---------------|
| **First View** | Runs `ship-commander tui` | Main Bridge appears, onboarding overlay: "Welcome to Ship Commander!" | "Whoa, what is all this?" | **Reassurance** - First-run tour explains panels |
| **Tour** | Presses `[Enter]` to advance tour | Overlay highlights Mission Control: "Agents (ensigns) execute tasks" | "Okay, agents are like helpers" | **Onboarding** - Simplified terminology |
| **Start Directive** | Runs `ship-commander start` | "Directive #42 started. Agent ensign-implementer dispatched." | "Hope I didn't break anything" | **Guidance** - Clear confirmation message |
| **Monitoring** | Navigates to Mission Control | "▸ impl-abc [d42] impl RED 00:00:15" | "It's working! Phase RED?" | **Curiosity** - Agent is visible, phase is unclear |
| **Phase Confusion** | Sees "RED 00:00:15" | Presses `[?]` for help → "RED: Writing failing test (TDD phase 1)" | "Oh, it's test-driven development" | **Learning** - Inline help explains concepts |
| **Progress** | Watches phase change | "d42 RED→GREEN PASS ✓" then "GREEN 00:00:30" | "It passed! Now implementing?" | **Excitement** - Progress is visible |
| **Gate Anxiety** | Sees "VERIFY_GREEN RUNNING..." | "Typecheck [████████] PASS, Lint [████████] PASS, Tests [███░░░] 30%" | "Are the tests passing?" | **Nervousness** - Gates are opaque |
| **Gate Pass** | Sees "VERIFY_GREEN PASS ✓" | "d42 GREEN→REFACTOR" | "Phew, it passed" | **Relief** - Gates are clear |
| **Completion** | Sees "d42 REVIEW APPROVED" | "Press [a] to approve merge or [h] to halt for review" | "Should I approve? What if it's wrong?" | **Fear** - Afraid to break production |
| **Safe Decision** | Presses `[h]` to halt | "Directive #42 halted. Worktree preserved at .ship-commander/worktrees/42" | "Safe, I can review it manually" | **Safety** - Halt doesn't destroy work |
| **Review** | Opens worktree in editor | Sees generated code, reviews tests | "This looks good actually" | **Confidence** - Work is inspectable |
| **Retry** | Runs `ship-commander start` again | "Directive #42 retried. Agent ensign-reviewer dispatched." | "Let's try again" | **Control** - Can restart safely |
| **Success** | Sees "d42 COMPLETED" | "✓ Directive #42 merged to main branch. Worktree cleaned up." | "I did it! 🎉" | **Achievement** - Celebration animation |

#### Key Emotional Touchpoints

1. **Onboarding Overlay**: "Welcome to Ship Commander!" → **"I'm being guided"**
2. **Inline Help**: `[?]` shows "RED: Writing failing test" → **"I'm learning"**
3. **Gate Progress**: "Tests [███░░░] 30%" → **"I see what's happening"**
4. **Safe Halt**: "Worktree preserved" → **"I can't break anything"**
5. **Celebration**: Success chime + checkmark → **"I accomplished something"**

#### Pain Points to Avoid

- **Information overload** (hide advanced panels until `[Tab]` to advanced mode)
- **Jargon without explanation** (tooltip on hover: "Directive = task to complete")
- **Destructive actions without confirmation** (halt asks: "Pause directive #42? Worktree will be preserved.")
- **No undo** (retry restores worktree, not destructive)

---

### Journey 3: Fleet Admiral - Multi-Project Wave Approval

**User**: Fleet Admiral (manager)
**Goal**: Approve completed wave across 3 projects without getting lost in details
**Expected Outcome**: Wave approved, merged, all projects updated

#### Emotional Arc

| Phase | User Action | TUI Response | Emotional State | Design Intent |
|-------|-------------|--------------|-----------------|---------------|
| **Executive View** | Runs `ship-commander tui --executive` | High-level dashboard: "Projects: auth (80%), ui (60%), infra (100%)" | "Good overview, I can see everything" | **Efficiency** - No clutter, just metrics |
| **Wave Alert** | Sees notification | "Wave 2 Complete (auth): 4 directives ready for merge" | "Time to approve" | **Timeliness** - Proactive notification |
| **Batch View** | Presses `[w]` for Wave Manager | Wave overview: "Wave 1 [████████] 4/4 merged, Wave 2 [████████] 4/4 ready" | "Wave 1 is done, Wave 2 is waiting" | **Clarity** - Wave status is obvious |
| **Conflict Check** | Reviews Wave 2 directives | "d6-d9: All gates passing ✓, No conflicts detected ✓" | "Clean merge, no manual work" | **Confidence** - Risks are surfaced |
| **Velocity Check** | Views timeline | "Wave 1 completed in 4h, Wave 2 in 3.5h (velocity +12%)" | "Team is speeding up" | **Insight** - Trends are visible |
| **Batch Approval** | Presses `[m]` to merge all ready | Confirmation: "Merge 4 directives to main branch?" | "Let's ship this" | **Control** - One action, clear impact |
| **Merge Progress** | Watches merge | "Merging d6... ✓, Merging d7... ✓, Merging d8... ✓, Merging d9... ✓" | "Smooth, no errors" | **Predictability** - Progress is linear |
| **Verification** | Sees post-merge gates | "Typecheck ✓, Lint ✓, Build ✓, Tests ✓" | "All green, ship it" | **Quality** - Gates give confidence |
| **Completion** | Sees summary | "✓ Merged 4/4 Wave 2 directives. Project auth: 100% complete." | "Done, moving on" | **Efficiency** - Clear outcome, no loose ends |
| **Project Switch** | Presses `[Tab]` to next project | Dashboard shows: "Projects: auth (100%), ui (60%), infra (100%)" | "Next up: UI project" | **Flow** - Easy navigation between projects |

#### Key Emotional Touchpoints

1. **Executive Dashboard**: One-line project summaries → **"I'm in control"**
2. **Wave Alert**: "Wave 2 Complete" notification → **"I'm informed"**
3. **Conflict Check**: "No conflicts detected" → **"No surprises"**
4. **Velocity Trend**: "+12% faster" → **"Team is improving"**
5. **Batch Approval**: One merge action for 4 directives → **"Efficient"**
6. **All Gates Passing**: "Typecheck ✓, Lint ✓, Build ✓, Tests ✓" → **"Quality is assured"**

#### Pain Points to Avoid

- **Detail overload** (executive mode hides individual gate results by default)
- **Unclear impact** (show what will merge: "4 directives, 12 files affected")
- **No rollback info** (show "Revert available for 5 minutes" after merge)
- **Slow navigation** ( `[Tab]` between projects is instant, no page reloads)

---

## Emotional Touchpoints

### When Should Users Feel Like a "Starfleet Admiral"?

**Admiral Mode** (strategic, command, oversight):

1. **Mission Briefing (Intake Start)**
   - **Visual**: "MISSION BRIEFING ROOM" header, LCARS orange border
   - **Audio**: (Optional) "Red alert" klaxon or "Ship's computer" voice
   - **Emotion**: "I'm initiating a mission"

2. **Specialist Dispatch (Parallel Execution)**
   - **Visual**: Three panels updating live, "WARP CORE ONLINE" header
   - **Metaphor**: "My away teams are beaming down"
   - **Emotion**: "I'm coordinating multiple experts"

3. **Plan Approval (Synthesis Complete)**
   - **Visual**: "COVERAGE REPORT: 15/15 UCs covered ✓"
   - **Metaphor**: "Mission parameters validated, engage"
   - **Emotion**: "I'm making a strategic decision"

4. **Wave Completion (Batch Merge)**
   - **Visual**: "Wave 2: [████████] 4/4 done" → "MERGE ALL READY?"
   - **Metaphor**: "Formation returned to base, debriefing complete"
   - **Emotion**: "My fleet executed successfully"

5. **Red Alert (Critical Failure)**
   - **Visual**: Pulsing red border, "HALTED: gate:lint FAIL"
   - **Audio**: (Optional) Red alert klaxon
   - **Emotion**: "I need to intervene personally"

### When Should Users Feel Like a "Developer"?

**Developer Mode** (tactical, debugging, learning):

1. **Gate Failure (Debugging)**
   - **Visual**: "Lint: 3 syntax errors in src/auth.ts:45-52"
   - **Emotion**: "I need to fix this code"
   - **Action**: Press `[s]` to open worktree in editor

2. **Agent Stuck (Troubleshooting)**
   - **Visual**: "cmdr-abc STATUS: STUCK phase=RED timeout=300s"
   - **Emotion**: "Why is it stuck? What's the error?"
   - **Action**: Press `[Enter]` for agent details, see last 10 lines of output

3. **Dependency Conflict (Planning)**
   - **Visual**: "Cycle detected: d1 → d2 → d1"
   - **Emotion**: "I need to fix the dependency graph"
   - **Action**: Edit PRD to clarify task ordering

4. **Worktree Inspection (Code Review)**
   - **Visual**: "Worktree: .ship-commander/worktrees/42"
   - **Emotion**: "Let me review what the agent produced"
   - **Action**: Open terminal/editor, review generated code

5. **Manual Retry (Recovery)**
   - **Visual**: "Retry directive #42? Worktree will be reset."
   - **Emotion**: "I'm recovering from a failure"
   - **Action**: Press `[y]` to retry with fresh worktree

### Emotional Progression by User Experience

| User Level | Week 1 | Week 2-4 | Week 5-8 | Week 9+ |
|------------|--------|----------|----------|---------|
| **Junior Lieutenant** | "Overwhelmed, what is all this?" | "Getting it, agents are helpers" | "Comfortable, can monitor workflows" | "Power user, tweaking configs" |
| **Staff Commander** | "Cool aesthetic, how does it work?" | "Efficient, love the parallel execution" | "Productive, shipping features fast" | "Expert, customizing roles and skills" |
| **Fleet Admiral** | "Interesting, but is it useful?" | "Helpful visibility into team velocity" | "Valuable, can spot blockers early" | "Essential, can't imagine managing without it" |

---

## Persona-Specific UI Needs

### Staff Commander (Power User)

**UI Requirements**:
1. **Full LCARS Theme**: Orange/blue/purple color scheme, box-drawing borders
2. **Real-Time Updates**: Agent panels refresh every 1s, gates every 2s
3. **Keyboard-Only Navigation**: Never need mouse, all actions have shortcuts
4. **Advanced Panels**: Dependency graph, wave summary, gate details (all visible)
5. **Log Streaming**: Engineering Logs panel with 8+ lines, scrollable
6. **Animation Controls**: Can enable/disable, adjust speed (fast/normal/slow)
7. **Multi-Project**: `[Tab]` between projects, show aggregated status

**Key Shortcuts**:
- `[a]` Focus Agents panel
- `[m]` Focus Mission Board
- `[t]` Focus Tactical Display
- `[l]` Focus Engineering Logs
- `[c]` Focus Command Interface
- `[Space]` Pause/resume propulsion
- `[h]` Halt directive
- `[r]` Retry directive
- `[w]` Wave Manager

**Settings Preferences**:
```json
{
  "theme": "lcars",
  "animations": true,
  "animationSpeed": "normal",
  "soundEffects": true,
  "logLines": 12,
  "refreshRate": 1000,
  "advancedMode": true,
  "mouseSupport": false
}
```

---

### Junior Lieutenant (Novice User)

**UI Requirements**:
1. **Simplified Theme**: Green/yellow/red status colors, minimal LCARS decoration
2. **Basic Mode**: Hide advanced panels (Dependency Graph, Wave Summary)
3. **Inline Help**: Contextual hints (`[?]` for help, tooltips on hover)
4. **Safe Defaults**: Can't halt without confirmation, worktrees preserved
5. **Progress Indicators**: Clear "what's happening now" labels
6. **Error Explanations**: Human-readable messages, not stack traces
7. **Onboarding**: First-run tour, tutorial mode

**Key Shortcuts**:
- `[?]` Show help overlay
- `[Enter]` Advance tour / Select item
- `[Tab]` Next panel (basic: Agents → Mission → Logs)
- `[Up/Down]` Navigate list
- `[q]` Quit TUI

**Settings Preferences**:
```json
{
  "theme": "simple",
  "animations": true,
  "animationSpeed": "slow",
  "soundEffects": false,
  "logLines": 4,
  "refreshRate": 2000,
  "advancedMode": false,
  "mouseSupport": true,
  "onboarding": true,
  "tutorialMode": true
}
```

---

### Fleet Admiral (Manager)

**UI Requirements**:
1. **Executive Theme**: Minimal decoration, status-focused colors (red/green)
2. **Executive Mode**: Hide details, show aggregated metrics only
3. **Multi-Project Dashboard**: All projects in one view
4. **Trend Visualization**: Velocity over time, common failure modes
5. **Alerting**: Notifications for critical issues only
6. **Drill-Down**: Click into details only when needed
7. **Reporting**: Export summaries for standups

**Key Shortcuts**:
- `[Tab]` Next project
- `[w]` Wave Manager (batch approval)
- `[e]` Export report (JSON/CSV)
- `[a]` Approve all ready
- `[q]` Quit TUI

**Settings Preferences**:
```json
{
  "theme": "minimal",
  "animations": false,
  "animationSpeed": "instant",
  "soundEffects": false,
  "logLines": 0,
  "refreshRate": 5000,
  "advancedMode": false,
  "mouseSupport": true,
  "executiveMode": true,
  "multiProject": true,
  "alerts": "critical_only"
}
```

---

## Design Implications

### Consistent Mental Model Across Personas

**Challenge**: The same TUI must serve novices (Junior Lieutenant) and experts (Staff Commander) without confusing either.

**Solution**: **Progressive Disclosure** + **Mode Toggling**

1. **Basic Mode** (default for first-time users):
   - Hide: Dependency Graph, Wave Summary, Gate Details
   - Show: Agents (simplified), Mission Board (counts only), Logs (last 3 lines)
   - Add: Inline help, contextual hints
   - Terminology: Hybrid ("Task #42" instead of "Directive #42")

2. **Advanced Mode** (toggle via `[Shift+A]` or settings):
   - Show: All panels, full details, keyboard shortcuts
   - Hide: Inline help, hints
   - Terminology: Full Star Trek ("Directive #42", "Ensign impl-abc")

3. **Executive Mode** (toggle via `--executive` flag or settings):
   - Hide: Individual agent panels, gate details, logs
   - Show: Aggregated metrics, trends, multi-project view
   - Terminology: Business language ("Feature X 80% complete")

### Adaptive UI Based on User Behavior

**Detect Persona** and adjust UI automatically:

| Behavior | Detected Persona | UI Adjustment |
|----------|------------------|---------------|
| First-time user (no `.ship-commander/` dir) | Junior Lieutenant | Show onboarding tour, enable basic mode |
| User runs `--executive` flag | Fleet Admiral | Hide details, show aggregated metrics |
| User uses `[h]`, `[r]`, `[w]` shortcuts frequently | Staff Commander | Show advanced panels, full LCARS theme |
| User never uses advanced features | Junior Lieutenant | Keep basic mode, hide advanced panels |
| User has 3+ projects | Fleet Admiral | Show multi-project dashboard |

### Terminology Translation Layer

**Problem**: "Directive" means different things to different personas.

**Solution**: **User-configurable terminology** in settings:

```json
{
  "terminology": {
    "directive": "directive",  // or "task", "story", "feature"
    "agent": "ensign",         // or "agent", "helper", "worker"
    "gate": "gate",            // or "check", "validation", "test"
    "wave": "wave",            // or "batch", "sprint", "iteration"
    "captain": "captain"       // or "operator", "user", "you"
  }
}
```

**TUI renders**: Display labels use user's preferred terminology throughout.

---

## Next Steps for Product Manager

1. **Validate Personas**: Do these three personas cover the actual user base? Any missing?
2. **Prioritize Features**: Which UI features are MVP vs. nice-to-have for each persona?
3. **Map Functional Groups**: How do UC-INTAKE, UC-ROLES, UC-ANALYZE, etc. map to user journeys?
4. **Define Success Metrics**: How do we measure if each persona is successful?

---

**END OF DOCUMENT**
