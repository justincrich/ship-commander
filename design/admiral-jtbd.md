# Admiral Jobs-to-Be-Done Analysis

**Date**: 2026-02-10
**Source**: PRD (`.spec/prd.md`), UX Design Plan (`design/UX-DESIGN-PLAN.md`), Technical Requirements (`.spec/technical-requirements.md`)
**Reframe**: User-provided revised mental model (Fleet, Directives, Agent Roster, Starships, Launching, Messages, Directive Planning)

---

## Table of Contents

1. [Revised Mental Model Mapping](#1-revised-mental-model-mapping)
2. [Admiral JTBD by Category](#2-admiral-jtbd-by-category)
3. [Four Forces Analysis](#3-four-forces-analysis)
4. [Complete Workflow Catalog](#4-complete-workflow-catalog)
5. [Screen Implications](#5-screen-implications)
6. [PRD Use Case Mapping](#6-prd-use-case-mapping)
7. [Gap Analysis](#7-gap-analysis)
8. [Job Priority Assessment](#8-job-priority-assessment)

---

## 1. Revised Mental Model Mapping

The user's revised mental model reframes the PRD's concepts into a more thematic, Admiral-centric vocabulary. This table maps each revised concept to the PRD equivalent and identifies what is preserved, what is renamed, and what is genuinely new.

| Revised Concept | PRD Equivalent | Relationship | Notes |
|----------------|---------------|-------------|-------|
| **Fleet** | No direct equivalent (implicit in project scope) | **NEW** | The PRD has no concept of multiple ships or fleet-level management. It assumes a single commission at a time. The revised model introduces a fleet as a collection of ships -- an explicit grouping layer above commissions. |
| **Starship (Ship)** | No direct equivalent | **NEW** | Ships are named groupings of agents. The PRD assigns agents per-mission, not per-ship. Ships add a persistent identity layer and a constraint: one directive at a time, but multiple concurrent missions within that directive. |
| **Directive** | Commission (PRD) | **RENAME + REFINE** | A directive is effectively a PRD/commission. The key refinement: a directive is assigned to a specific ship, creating a ship-to-directive binding that the PRD does not model. |
| **Agent Roster** | Roles (Captain, Commander, etc.) in PRD | **EXPAND** | The PRD defines roles with fixed responsibilities. The revised model makes agents first-class entities with customizable attributes: name, age, backstory, skills (harness skill strings), model (validated against harness), mission prompt. This is a significant expansion -- agents become persistent, configurable resources rather than ephemeral role-fillers. |
| **Commissioning Starships** | N/A | **NEW** | Naming and creating ships. The PRD has no ship creation ceremony. This adds a fun, thematic onboarding step. |
| **Launching Ships** | `sc3 execute` / Commander Execution | **RENAME + REFINE** | "Launching" a ship means starting execution of a directive. "Opening" a ship means viewing its status. Maps to the execution flow but with ship identity attached. |
| **Messages/Questions** | UC-COMM-08 (Agent to Admiral question) | **PRESERVE** | Direct mapping. Messages from captains and commanders to the Admiral during planning and execution. |
| **Directive Planning** | Captain's Ready Room (UC-COMM-02 through UC-COMM-07) | **RENAME + EXPAND** | Planning happens in the "captain's ready room" with the captain, commander, and additional crew (design officer). The revised model preserves this but frames it as "directive planning" with explicit crew member participation. |
| **Commanders** | Commander role in PRD | **PRESERVE** | Mission execution authority. |
| **Captains** | Captain role in PRD | **PRESERVE** | Functional requirement authority. |
| **Ensigns (Specialists)** | Implementer, Reviewer, Design Officer in PRD | **RENAME** | Specialist agents that execute specific missions. |
| **Crew** | Agent instances assigned to missions | **RENAME** | The collective agents assigned to a ship. |
| **Missions** | Missions in PRD | **PRESERVE** | Atomic coding tasks within a directive. |

### Key Conceptual Additions in Revised Model

1. **Ship as persistent identity**: Ships exist beyond a single directive. They can be re-assigned. They have names, which means the Admiral has an emotional relationship with them.
2. **Agent Roster as configuration surface**: Agents are not just roles -- they are entities the Admiral creates, customizes, and maintains. This makes agent management a first-class job.
3. **Fleet as portfolio view**: Multiple ships on multiple directives simultaneously. The PRD's Executive Mode hints at this ("3 commissions active") but the revised model makes it explicit.
4. **Constraint: one directive per ship**: This is a design choice that prevents confusion. A ship focuses on one PRD at a time, but missions within that directive can run in parallel.

---

## 2. Admiral JTBD by Category

### Category A: Fleet Management

Jobs related to creating, organizing, and maintaining the Admiral's collection of ships and crews.

**JOB A.1: Commission a new starship**
- Job Statement: "When I want to expand my fleet's capacity, I want to create and name a new starship, so I can assign it directives and put it to work."
- Dimensions:
  - Functional: Create a named entity that groups agents and can receive directive assignments
  - Emotional: Pride in naming, sense of building something, fleet commander identity
  - Social: Sharing fleet names with peers, having a recognizable stable of ships
- Frequency: Low (fleet grows slowly)
- Criticality: Required (no ships = no work gets done)

**JOB A.2: View fleet status**
- Job Statement: "When I have multiple ships working on different directives, I want to see the overall fleet status at a glance, so I can identify which ships need attention and where progress is happening."
- Dimensions:
  - Functional: Aggregate view of all ships, their directives, and completion percentages
  - Emotional: Feeling of control, confidence that work is progressing
  - Social: Ability to report status to stakeholders
- Frequency: High (multiple times per session)
- Criticality: High (without this, the Admiral is blind)

**JOB A.3: Assign a directive to a ship**
- Job Statement: "When I have a PRD ready and a ship available, I want to assign the directive to a specific ship, so that ship's crew can begin planning and execution."
- Dimensions:
  - Functional: Bind a directive (PRD) to a ship, making it the ship's active mission
  - Emotional: Sense of deployment, "sending the ship on a mission"
  - Social: N/A
- Frequency: Medium (when new directives arrive or ships complete)
- Criticality: Required (the bridge between planning and doing)

**JOB A.4: Decommission or reassign a ship**
- Job Statement: "When a ship is no longer needed or I want to reorganize my fleet, I want to decommission a ship or reassign its crew, so I can optimize my resources."
- Dimensions:
  - Functional: Remove a ship or redistribute its agents
  - Emotional: Bittersweet if attached to the name; satisfaction if optimizing
  - Social: N/A
- Frequency: Low
- Criticality: Low (nice-to-have for fleet hygiene)

---

### Category B: Agent Roster Management

Jobs related to creating, configuring, and maintaining the agents that crew the ships.

**JOB B.1: Create a new agent**
- Job Statement: "When I need a specialist for a specific role, I want to create an agent with a name, role, skills, model, and mission prompt, so I can staff my ships with purpose-built crew members."
- Dimensions:
  - Functional: Define agent identity (name, role), capabilities (skills as harness skill strings), runtime config (model validated against harness), and behavior (mission prompt)
  - Emotional: Creative satisfaction, feeling of crafting a team
  - Social: Sharing agent configurations with other users
- Frequency: Medium (at fleet setup and when new specializations needed)
- Criticality: Required (no agents = no one to do the work)

**JOB B.2: Configure agent capabilities**
- Job Statement: "When I want to tune how an agent performs, I want to update its skills, model, or mission prompt, so the agent is optimized for the work it does."
- Dimensions:
  - Functional: Edit agent properties: skill strings (validated on target harness), model selection (validated as permissible on harness), mission prompt text
  - Emotional: Tinkering satisfaction, feeling of control over AI behavior
  - Social: N/A
- Frequency: Medium (iterative tuning)
- Criticality: High (misconfigured agents waste time and tokens)

**JOB B.3: Add optional personality to an agent**
- Job Statement: "When I want my agents to feel distinct and memorable, I want to optionally set an age and backstory for each agent, so my fleet has character even though these traits do not affect dev work."
- Dimensions:
  - Functional: Set optional fields (age, backstory) that are stored but not injected into dev prompts
  - Emotional: Fun, personalization, attachment to the fleet
  - Social: Sharing agent "bios" with colleagues
- Frequency: Low (one-time per agent)
- Criticality: None (purely fun, zero functional impact)

**JOB B.4: Assign agents to a ship's crew**
- Job Statement: "When I commission a ship, I want to assign specific agents to crew it, so the right specialists are on board for the directive."
- Dimensions:
  - Functional: Bind agents from the roster to a ship's crew manifest, ensuring required roles (Captain, Commander) are filled
  - Emotional: Sense of assembling a team
  - Social: N/A
- Frequency: Medium (per ship commissioning or crew rotation)
- Criticality: Required (ships need crew)

**JOB B.5: Review agent performance history**
- Job Statement: "When I want to understand how well an agent is performing, I want to see its history of missions, gate results, and revision counts, so I can decide whether to retune or replace it."
- Dimensions:
  - Functional: Aggregate metrics per agent: missions completed, gate pass rates, average revisions
  - Emotional: Informed confidence in agent choices
  - Social: N/A
- Frequency: Low to Medium
- Criticality: Medium (useful but not blocking)

---

### Category C: Directive Management

Jobs related to creating, defining, and tracking PRD-level initiatives.

**JOB C.1: Create a directive from a PRD**
- Job Statement: "When I have a product requirements document, I want to load it as a directive with parsed use cases and acceptance criteria, so the system can decompose it into missions."
- Dimensions:
  - Functional: Parse a Markdown PRD file into a structured directive with use cases, ACs, scope
  - Emotional: Feeling of kicking off something important
  - Social: N/A
- Frequency: Medium (each new initiative)
- Criticality: Required (the starting point of all work)
- PRD Mapping: UC-COMM-01

**JOB C.2: Review directive status**
- Job Statement: "When a directive is in progress, I want to see which missions are done, which are active, and which are blocked, so I know how close we are to completion."
- Dimensions:
  - Functional: Directive-level dashboard showing mission breakdown, wave progress, use-case coverage
  - Emotional: Progress gratification or urgency awareness
  - Social: Status reporting to stakeholders
- Frequency: High
- Criticality: High
- PRD Mapping: UC-EXEC-10, UC-TUI-04, UC-TUI-09

**JOB C.3: Shelve a directive for later**
- Job Statement: "When priorities shift, I want to shelve an in-progress directive so its plan is saved and I can resume later without re-planning."
- Dimensions:
  - Functional: Persist full plan state (missions, sign-offs, messages, iterations) and pause execution
  - Emotional: Relief that work is not lost; anxiety about context loss
  - Social: N/A
- Frequency: Low
- Criticality: Medium
- PRD Mapping: UC-COMM-10

**JOB C.4: Resume a shelved directive**
- Job Statement: "When I am ready to return to a previously shelved directive, I want to resume it from where I left off, so I do not repeat planning work."
- Dimensions:
  - Functional: Re-load plan state, re-spawn agent sessions, continue from last state
  - Emotional: Confidence that the system remembers
  - Social: N/A
- Frequency: Low
- Criticality: Medium
- PRD Mapping: UC-COMM-10

---

### Category D: Directive Planning (Ready Room)

Jobs related to the collaborative planning loop where the directive is decomposed into executable missions.

**JOB D.1: Initiate directive planning**
- Job Statement: "When a directive is assigned to a ship, I want to kick off the planning process in the captain's ready room, so the ship's officers can decompose the directive into missions."
- Dimensions:
  - Functional: Spawn planning sessions for Captain, Commander, and any additional crew (Design Officer); inject directive context
  - Emotional: Anticipation, feeling of "things starting"
  - Social: N/A
- Frequency: Medium (once per directive)
- Criticality: Required (no planning = no execution)
- PRD Mapping: UC-COMM-02, UC-COMM-03, UC-COMM-04, UC-COMM-05

**JOB D.2: Observe planning progress**
- Job Statement: "When planning is underway, I want to see which agents have completed their analysis, what the sign-off status is, and how many iterations have occurred, so I know when to expect a plan for review."
- Dimensions:
  - Functional: Planning dashboard showing agent session status, iteration count, sign-off matrix, pending questions
  - Emotional: Patience vs. impatience; confidence vs. anxiety
  - Social: N/A
- Frequency: High during planning
- Criticality: High
- PRD Mapping: UC-TUI-01

**JOB D.3: Answer agent questions during planning**
- Job Statement: "When an agent encounters ambiguity during planning, I want to receive and answer their question, so planning can proceed with correct information."
- Dimensions:
  - Functional: Modal presents question with context, domain, options; Admiral selects answer or provides free text; answer routed back
  - Emotional: Feeling needed, authority exercised; mild annoyance if questions are trivial
  - Social: N/A
- Frequency: Variable (0-N per planning session)
- Criticality: High (planning blocks without answers)
- PRD Mapping: UC-COMM-08, UC-TUI-02

**JOB D.4: Review the mission manifest**
- Job Statement: "When the ready room reaches consensus, I want to review the full mission manifest with sequencing, use-case coverage, and agent sign-offs, so I can approve, give feedback, or shelve the plan."
- Dimensions:
  - Functional: Overlay showing Glamour-rendered manifest, coverage summary, approve/feedback/shelve controls
  - Emotional: Judgment moment -- feeling of authority and responsibility
  - Social: N/A
- Frequency: Once per planning cycle (possibly repeated if feedback given)
- Criticality: Required (Admiral approval is a non-negotiable gate)
- PRD Mapping: UC-COMM-09, UC-TUI-03

**JOB D.5: Provide feedback to reconvene planning**
- Job Statement: "When the mission manifest is not satisfactory, I want to provide feedback that is injected back into the planning loop, so the agents can revise the plan."
- Dimensions:
  - Functional: Multi-line text input; feedback distributed to all agent sessions; planning loop reconvenes
  - Emotional: Constructive authority; frustration if repeated
  - Social: N/A
- Frequency: Low to Medium
- Criticality: High (the feedback loop)
- PRD Mapping: UC-COMM-09

**JOB D.6: Approve the plan and launch**
- Job Statement: "When I am satisfied with the mission manifest, I want to approve it, so the ship can begin executing missions."
- Dimensions:
  - Functional: Approve action transitions directive from planning to executing; Commander begins dispatch
  - Emotional: Commitment, excitement, "let's go"
  - Social: N/A
- Frequency: Once per directive
- Criticality: Required (nothing executes without approval)
- PRD Mapping: UC-COMM-09

---

### Category E: Ship Launch and Execution Monitoring

Jobs related to watching ships execute their directives and intervening when needed.

**JOB E.1: Launch a ship on its directive**
- Job Statement: "When the plan is approved, I want to launch the ship so it begins executing missions in wave order."
- Dimensions:
  - Functional: Trigger execution; Commander sequences missions, creates worktrees, dispatches agents
  - Emotional: Moment of launch -- excitement, anticipation
  - Social: N/A
- Frequency: Once per directive
- Criticality: Required
- PRD Mapping: UC-EXEC-01

**JOB E.2: Monitor active mission execution**
- Job Statement: "When missions are running, I want to see per-mission TDD phase progress, gate results, and agent status, so I can understand what is happening in real time."
- Dimensions:
  - Functional: Execution dashboard with agent grid, mission board, phase tracker, wave view, event log
  - Emotional: Watching the machine work -- satisfaction, vigilance
  - Social: N/A
- Frequency: High (continuous during execution)
- Criticality: High
- PRD Mapping: UC-TUI-04, UC-TUI-05, UC-TUI-06, UC-TUI-09

**JOB E.3: Open a ship to check directive progress**
- Job Statement: "When I want to check on a specific ship, I want to open its view and see the status of its directive, missions, and agents, so I can focus on one ship at a time."
- Dimensions:
  - Functional: Navigate from fleet view to ship-specific execution dashboard
  - Emotional: Focused attention
  - Social: N/A
- Frequency: High
- Criticality: High
- PRD Mapping: UC-TUI-04 (implied drill-down from Executive mode)

**JOB E.4: Halt a stuck or failing mission**
- Job Statement: "When a mission is stuck, looping, or producing bad results, I want to halt it, so resources are not wasted."
- Dimensions:
  - Functional: Halt command via command bar or keyboard shortcut; confirmation dialog; mission returns to backlog or is marked halted
  - Emotional: Decisive intervention, frustration at the problem
  - Social: N/A
- Frequency: Low to Medium
- Criticality: High (resource protection)
- PRD Mapping: UC-TUI-07, UC-EXEC-06

**JOB E.5: Retry a halted mission**
- Job Statement: "When a halted mission can be reattempted, I want to retry it, so it gets another chance at completion."
- Dimensions:
  - Functional: Retry command resets revision count, returns mission to backlog
  - Emotional: Optimism, "one more try"
  - Social: N/A
- Frequency: Low
- Criticality: Medium
- PRD Mapping: UC-TUI-07

**JOB E.6: Answer agent questions during execution**
- Job Statement: "When a Commander or Captain surfaces a question during execution, I want to answer it so the mission can proceed."
- Dimensions:
  - Functional: Same modal as planning questions; answer routed to asking agent
  - Emotional: Urgency (execution is blocked)
  - Social: N/A
- Frequency: Variable
- Criticality: High (execution blocks without answers)
- PRD Mapping: UC-COMM-08, UC-TUI-02

**JOB E.7: Review demo tokens / proof artifacts**
- Job Statement: "When a mission completes, I want to review the demo token to verify the work without reading all the code, so I can confirm quality."
- Dimensions:
  - Functional: View rendered demo token (Glamour markdown) with evidence sections, test results, diff refs
  - Emotional: Trust but verify; satisfaction at completed work
  - Social: N/A
- Frequency: Once per completed mission
- Criticality: Medium (Commander auto-validates, but Admiral may want to inspect)
- PRD Mapping: UC-EXEC-07, Mission Detail overlay

**JOB E.8: Adjust WIP limits**
- Job Statement: "When I want more or fewer concurrent agents, I want to change the WIP limit, so I can balance speed against resource usage."
- Dimensions:
  - Functional: `wip <n>` command; immediate effect on next dispatch cycle
  - Emotional: Control over pacing
  - Social: N/A
- Frequency: Low
- Criticality: Low
- PRD Mapping: UC-TUI-07

---

### Category F: Messages and Communications

Jobs related to the Admiral's communication channels with ship officers.

**JOB F.1: Receive and respond to officer questions**
- Job Statement: "When a captain or commander needs my input, I want to see their question with full context and respond quickly, so the ship is not blocked waiting for me."
- Dimensions:
  - Functional: Question modal with agent identity, domain, context, options/free-text, broadcast toggle
  - Emotional: Responsibility, importance, mild pressure
  - Social: N/A
- Frequency: Variable (0-N per session)
- Criticality: High (blocking)
- PRD Mapping: UC-COMM-08, UC-TUI-02

**JOB F.2: Review communication history**
- Job Statement: "When I want to recall what was discussed during planning, I want to see the full message log between agents and my Q&A history, so I can understand how decisions were made."
- Dimensions:
  - Functional: Scrollable message log, question/answer pairs linked by ID, planning audit trail
  - Emotional: Clarity, institutional memory
  - Social: N/A
- Frequency: Low to Medium
- Criticality: Medium (auditability)
- PRD Mapping: UC-TUI-01 (inter-agent message log), UC-COMM-06, UC-STATE-09

**JOB F.3: Broadcast a directive to all agents**
- Job Statement: "When I have information that all agents need, I want to broadcast a message, so everyone is aligned."
- Dimensions:
  - Functional: Option to broadcast an answer to all sessions, not just the asking agent
  - Emotional: Authority, clarity of communication
  - Social: N/A
- Frequency: Low
- Criticality: Low
- PRD Mapping: UC-COMM-08 (broadcast toggle on question answers)

---

### Category G: System Health and Observability

Jobs related to understanding whether the system itself is functioning correctly.

**JOB G.1: Monitor system health**
- Job Statement: "When ships are running, I want to see system health indicators (WIP utilization, stuck agents, orphaned missions, Doctor heartbeat), so I can trust the system is working."
- Dimensions:
  - Functional: Health panel with WIP bar, Doctor status, stuck/orphan counts
  - Emotional: Trust in the system, anxiety when things degrade
  - Social: N/A
- Frequency: Periodic glances during execution
- Criticality: Medium
- PRD Mapping: UC-TUI-08

**JOB G.2: Investigate stuck or dead agents**
- Job Statement: "When the system reports a stuck or dead agent, I want to drill into the agent's details to understand what happened, so I can decide whether to retry or reassign."
- Dimensions:
  - Functional: Agent detail overlay with output, phase, health, elapsed time
  - Emotional: Diagnostic mode, problem-solving
  - Social: N/A
- Frequency: Low (only on problems)
- Criticality: High (when it happens, it is urgent)
- PRD Mapping: UC-TUI-05 (Agent Detail overlay)

**JOB G.3: View event log for debugging**
- Job Statement: "When I need to understand what happened, I want to scroll through the event log with severity filtering, so I can trace the sequence of events."
- Dimensions:
  - Functional: Scrollable, filterable event log with structured entries
  - Emotional: Detective mode, understanding
  - Social: N/A
- Frequency: Low to Medium
- Criticality: Medium
- PRD Mapping: UC-TUI-06

---

### Category H: Configuration and Setup

Jobs related to initial system setup and ongoing configuration.

**JOB H.1: Initialize Ship Commander for a project**
- Job Statement: "When I start using Ship Commander on a new codebase, I want to initialize it with project config (gate commands, harness defaults, WIP limits), so the system is tuned for my project."
- Dimensions:
  - Functional: `sc3 init` creates config file (TOML), initializes Beads, validates dependencies
  - Emotional: Setup friction vs. "things just work"
  - Social: N/A
- Frequency: Once per project
- Criticality: Required (one-time gate)
- PRD Mapping: UC-STATE-01, Constraints (config format)

**JOB H.2: Configure role-to-model mapping**
- Job Statement: "When I want to optimize cost vs. quality, I want to assign specific AI models to specific roles (Captain=Opus, Implementer=Sonnet), so I spend tokens wisely."
- Dimensions:
  - Functional: TOML config `[roles.captain] model = "opus"` with harness validation
  - Emotional: Cost awareness, optimization satisfaction
  - Social: N/A
- Frequency: Low (initial setup + occasional tuning)
- Criticality: Medium
- PRD Mapping: UC-HARN-08

**JOB H.3: Validate harness availability**
- Job Statement: "When I start Ship Commander, I want to know which harnesses (Claude Code, Codex) are available, so I can configure my fleet accordingly."
- Dimensions:
  - Functional: Startup check for `claude`, `codex`, `tmux`, `bd` on PATH
  - Emotional: Confidence that things will work
  - Social: N/A
- Frequency: Once per startup
- Criticality: Required (fail-fast)
- PRD Mapping: UC-HARN-07

---

## 3. Four Forces Analysis

The Four Forces framework analyzes the push, pull, anxiety, and habit forces that affect the Admiral's switching behavior.

### Forces Acting on the Admiral

**PUSH (away from current situation)**
- AI coding workflows degrade past 1-2 agents -- lost productivity
- Agents self-certify with vanity tests -- quality illusion
- No planning discipline -- agents jump to code, build the wrong thing
- No termination guarantee -- infinite loops waste tokens and time
- Collapsed authority -- no separation between planning, orchestration, and execution
- Human cannot verify what agents built without reading all code

**PULL (toward Ship Commander)**
- Parallel agent execution with deterministic oversight
- Collaborative planning before execution (Ready Room)
- Independent verification gates (no self-certification)
- Demo tokens as human-reviewable proof of work
- Persistent state that survives crashes
- Fun Star Trek theme that makes the work enjoyable
- Fleet model gives sense of command and control
- Agent Roster allows crafting purpose-built specialists

**ANXIETY (about adopting Ship Commander)**
- Learning curve: new concepts (directives, ships, missions, gates)
- Will it actually save time or add orchestration overhead?
- Token cost of running multiple agents (Captain, Commander, Design Officer, Implementer)
- Trust: will the Commander actually catch vanity tests?
- Configuration complexity (TOML, roles, models, gate commands)
- What happens when things go wrong? (crash recovery, stuck agents)
- Dependence on tmux, Beads, and specific CLI tools being on PATH

**HABIT (keeping current behavior)**
- "I just run Claude Code manually and it works fine for small tasks"
- Existing muscle memory with single-agent workflows
- Direct control over what the agent does at each step
- No learning curve with current approach
- "I do not need parallel agents for my project size"

### Anxiety Mitigation Strategy (from UX Design Plan)

| Anxiety | Mitigation |
|---------|-----------|
| Learning curve | Basic mode with simplified terminology (Task, Batch, Check, Helper) and onboarding overlay |
| Orchestration overhead | Ready Room handles planning automatically; Admiral only approves |
| Trust in gates | Gate results shown in real time with full output; deterministic = verifiable |
| Configuration complexity | Sensible defaults; only `sc3 init` required to start |
| Crash recovery | Beads persists everything; restart reconstructs state in <10s |

---

## 4. Complete Workflow Catalog

### Workflow 1: First-Time Setup

**Trigger**: Admiral wants to use Ship Commander on a new project.

```
1. Admiral runs `sc3 init`
   - System checks PATH for: claude, codex (optional), tmux, bd
   - System creates sc3.toml with defaults
   - System runs `bd init` to create .beads/ directory
   - System reports available harnesses

2. Admiral edits sc3.toml (optional)
   - Configure gate commands for project (test, typecheck, lint, build)
   - Configure role-to-model mapping
   - Set WIP limit and max revisions

3. System ready for fleet operations
```

**Screen Implications**:
- No TUI needed -- CLI output only
- Could benefit from an interactive `sc3 init` wizard (Huh form in CLI mode)

---

### Workflow 2: Build the Agent Roster (NEW -- not in PRD)

**Trigger**: Admiral wants to create and configure the agents that will crew ships.

```
1. Admiral creates agents via CLI or TUI
   - Required: name, role (Captain/Commander/Ensign-specialist)
   - Required: skills (harness skill string, e.g., "backend-implement,go-engineer")
   - Required: model (e.g., "opus", "sonnet" -- validated against harness)
   - Required: mission prompt (the system prompt for this agent in dev work)
   - Optional: age, backstory (fun/flavor, not used in dev prompts)

2. System validates agent configuration
   - Model is permissible on configured harness
   - Skills are valid skill strings on target harness
   - Role is a recognized role type

3. Agent stored in roster (Beads or config)
   - Agent available for ship crew assignment

4. Admiral can later edit, clone, or retire agents
```

**Screen Implications**:
- **Agent Roster Screen** (NEW): List of all agents with role, model, skills summary
- **Agent Creation Form** (NEW): Huh form with fields for name, role, skills, model, mission prompt, optional age/backstory
- **Agent Detail/Edit** (NEW): View and modify agent configuration
- This is a significant UI addition not in the PRD or UX Design Plan

---

### Workflow 3: Commission a Starship (NEW -- not in PRD)

**Trigger**: Admiral wants to create a new ship in the fleet.

```
1. Admiral initiates ship creation
   - Names the ship (required, must be unique)
   - Optionally provides ship class or description

2. Admiral assigns crew from roster
   - Must assign a Captain (one per ship)
   - Must assign a Commander (one per ship)
   - Optionally assigns Ensigns (specialists: Design Officer, Implementer, Reviewer)
   - System validates required roles are filled

3. Ship is commissioned and available
   - Ship appears in fleet view
   - Ship status: "docked" (no active directive)
```

**Screen Implications**:
- **Ship Commissioning Form** (NEW): Huh form for ship name, crew assignment (multi-select from roster)
- **Fleet View** (NEW or expanded Executive Mode): Shows all ships with status
- Potential for a fun "launch ceremony" animation

---

### Workflow 4: Assign a Directive to a Ship

**Trigger**: Admiral has a PRD and a ship ready to work on it.

```
1. Admiral creates directive from PRD file
   - `sc3 directive create <prd-file>`
   - System parses Markdown into structured directive (use cases, ACs)
   - Directive persisted to Beads

2. Admiral assigns directive to a ship
   - `sc3 directive assign <directive-id> <ship-name>`
   - System validates ship has no active directive (one at a time)
   - System validates ship has required roles (Captain, Commander)
   - Directive bound to ship

3. Ship status changes to "planning"
   - Ship's officers notified (sessions will be spawned when planning begins)
```

**Screen Implications**:
- **Directive List Screen** (NEW or mapped to commission list): Shows all directives with status
- **Assignment Flow**: Could be CLI command or TUI interaction (select ship, select directive)

---

### Workflow 5: Directive Planning in the Ready Room

**Trigger**: Admiral initiates planning for a ship's assigned directive.

```
1. Admiral opens the Ready Room for a ship
   - `sc3 plan <ship-name>` or navigate to ship > Start Planning
   - TUI switches to Ready Room view

2. System spawns planning sessions
   - Captain session with directive context
   - Commander session with directive context
   - Design Officer session (if crew includes one) with directive context
   - Each session operates in isolated context

3. Agents analyze the directive (Admiral observes)
   - Captain: functional requirements, use-case-to-mission mapping
   - Commander: technical decomposition, mission sequencing
   - Design Officer: design requirements, UX implications
   - Inter-agent messages routed through structured message passing
   - Planning dashboard shows progress, iteration count, sign-offs

4. Agent questions surface to Admiral
   - Modal appears: "ADMIRAL -- QUESTION FROM CAPTAIN"
   - Admiral reads context, selects option or types answer
   - Optionally broadcasts answer to all agents
   - Answer routed back; planning resumes

5. Consensus check (deterministic)
   - All missions signed off by all required agents
   - All directive use cases covered by at least one mission
   - If consensus fails: loop iterates (up to max iterations)

6. Mission manifest presented to Admiral
   - Plan Review overlay with Glamour-rendered manifest
   - Use-case coverage summary
   - Admiral decision: Approve / Feedback / Shelve

7a. APPROVE: Directive transitions to approved
    - Ship ready for launch

7b. FEEDBACK: Admiral provides revision notes
    - Feedback injected into all agent sessions
    - Planning loop reconvenes
    - Return to step 3

7c. SHELVE: Plan saved for later
    - Full state persisted (missions, messages, sign-offs, iterations)
    - Ship returns to docked status
```

**Screen Implications**:
- **Ready Room Dashboard** (existing in PRD/UX): Agent sessions, message log, sign-off matrix, commission summary, question queue
- **Admiral Question Modal** (existing): Huh form with Select, Input, Confirm
- **Plan Review Overlay** (existing): Glamour markdown + approve/feedback/shelve
- Ship identity should be visible in the Ready Room header (e.g., "USS Enterprise -- Ready Room")

---

### Workflow 6: Launch a Ship

**Trigger**: Admiral has an approved plan and wants to start execution.

```
1. Admiral launches the ship
   - `sc3 launch <ship-name>` or Approve action from Plan Review
   - Directive transitions from approved to executing
   - TUI transitions from Ready Room to Main Bridge

2. Commander begins execution
   - Reads approved mission manifest
   - Computes wave execution order from dependency graph
   - Acquires surface-area locks
   - Dispatches agents to worktrees for Wave 1 missions

3. Admiral monitors from the bridge
   - Agent Grid shows active agents
   - Mission Board shows mission status (Kanban)
   - Phase Tracker shows per-AC TDD progress
   - Wave View shows parallel execution progress
   - Event Log streams state transitions and gate results

4. Execution proceeds autonomously
   - Commander runs verification gates
   - Commander enforces termination rules
   - Commander validates demo tokens
   - Waves advance as dependencies clear
```

**Screen Implications**:
- **Main Bridge** (existing in PRD/UX): The primary execution monitoring screen
- Ship identity in header (e.g., "USS Enterprise -- Main Bridge")
- Transition animation from Ready Room to Main Bridge

---

### Workflow 7: Respond to Questions During Execution

**Trigger**: A Captain or Commander surfaces a question while the ship is executing.

```
1. Agent emits question event
2. TUI shows Admiral Question Modal (interrupts current view)
3. Admiral reads context and responds
4. Answer routed to agent; execution resumes
```

**Screen Implications**:
- **Admiral Question Modal** (existing): Same modal as planning, overlaid on Main Bridge
- Terminal bell notification to get attention

---

### Workflow 8: Intervene on a Mission

**Trigger**: Admiral sees a mission that is stuck, failing repeatedly, or needs manual action.

```
1. Admiral identifies problem
   - Stuck agent warning in System Health
   - Mission with high revision count in Mission Board
   - Red alert animation on critical failure

2. Admiral investigates
   - Drill into Mission Detail (Enter on mission)
   - View gate evidence, AC progress, revision history
   - Drill into Agent Detail (Enter on agent)
   - View agent output, phase, health

3. Admiral takes action
   - HALT: Stops the mission (with confirmation)
   - RETRY: Returns mission to backlog for another attempt
   - WIP adjustment: Change concurrent agent count

4. Commander processes the action
   - Halted mission's locks released
   - Retried mission enters next dispatch cycle
```

**Screen Implications**:
- **Mission Detail Overlay** (existing): Deep view of mission ACs, gates, demo token
- **Agent Detail Overlay** (existing): Agent output, phase, health
- **Confirm Dialog** (existing): Huh Confirm for destructive actions
- **Command Bar** (existing): `halt <id>`, `retry <id>`, `wip <n>`

---

### Workflow 9: Review Completed Directive

**Trigger**: All missions in a directive are done.

```
1. Commander reports directive completion
2. Directive transitions to completed
3. Admiral reviews results
   - All demo tokens available for review
   - Use-case coverage confirmed at 100%
   - Mission completion log shows all gate evidence

4. Ship returns to docked status
   - Ship available for next directive assignment
```

**Screen Implications**:
- **Directive Completion Summary** (partially covered by existing PRD): Could be a new overlay showing all demo tokens, coverage, and metrics
- Ship status update in Fleet View

---

### Workflow 10: Fleet-Level Oversight

**Trigger**: Admiral has multiple ships on multiple directives and wants the big picture.

```
1. Admiral opens Fleet View (Executive Mode)
   - All ships listed with: ship name, directive title, progress %, health
   - Velocity metrics and trend lines
   - Blocker summary

2. Admiral identifies ships needing attention
   - Stuck ships highlighted
   - Ships with high revision counts flagged

3. Admiral drills into specific ship
   - Enter on ship opens its Main Bridge view
   - Can switch between ships

4. Admiral takes fleet-level actions
   - Approve pending plans across ships
   - Adjust WIP limits globally
```

**Screen Implications**:
- **Fleet View** (NEW or heavily expanded Executive Mode): The PRD's Executive Mode shows "3 commissions active" but the revised model makes this a ship-centric view with ship names, not just commission names
- Ship-level drill-down navigation

---

### Workflow 11: Manage Agent Roster Over Time (NEW -- not in PRD)

**Trigger**: Admiral wants to tune, add, or retire agents.

```
1. Admiral opens Agent Roster
   - See all agents with: name, role, model, skills, active assignment

2. Admiral edits an agent
   - Update skills, model, mission prompt
   - System re-validates against harness

3. Admiral creates a new specialist
   - For a new type of work (e.g., database specialist ensign)

4. Admiral retires an agent
   - Agent removed from active roster
   - Historical performance data retained
```

**Screen Implications**:
- **Agent Roster Screen** (NEW): Full agent management interface
- **Agent Editor** (NEW): Huh forms for editing agent properties

---

## 5. Screen Implications

### Complete Screen Inventory (Revised Model)

| Screen | Status vs PRD | Type | Purpose |
|--------|--------------|------|---------|
| **Fleet View** | NEW | primary | All ships, their directives, health, progress |
| **Agent Roster** | NEW | primary or overlay | All agents with config, performance, assignment |
| **Ship Commissioning** | NEW | form/wizard | Name ship, assign crew from roster |
| **Agent Creation/Edit** | NEW | form | Create or modify agent properties |
| **Directive List** | NEW (implied) | primary or panel | All directives with status |
| **Ready Room** | EXISTS (PRD) | primary | Planning dashboard for a specific ship |
| **Main Bridge** | EXISTS (PRD) | primary | Execution dashboard for a specific ship |
| **Admiral Question Modal** | EXISTS (PRD) | modal | Agent question with structured response |
| **Plan Review Overlay** | EXISTS (PRD) | overlay | Mission manifest review |
| **Mission Detail** | EXISTS (PRD) | overlay | Single mission deep view |
| **Agent Detail** | EXISTS (PRD) | overlay | Single agent deep view |
| **Help Overlay** | EXISTS (PRD) | overlay | Contextual keyboard shortcuts |
| **Confirm Dialog** | EXISTS (PRD) | modal | Destructive action confirmation |
| **Directive Completion Summary** | NEW (implied) | overlay | Post-completion review of all demo tokens |

### Navigation Model Update

The revised model requires a new top-level navigation structure:

```
sc3 tui
  |
  +-- [Fleet View] (NEW top-level)
  |     |-- Ship cards (name, directive, progress, health)
  |     |-- Fleet-level metrics
  |     +-- [Enter] on ship -> ship context
  |
  +-- [Ship Context] (when a ship is selected)
  |     |
  |     +-- [Ready Room] (if ship is planning)
  |     +-- [Main Bridge] (if ship is executing)
  |     +-- [Ship Detail] (crew, directive, history)
  |
  +-- [Agent Roster] (accessible from Fleet View)
  |     |-- Agent list with config summary
  |     |-- [Enter] on agent -> Agent Edit form
  |     +-- [n] new agent -> Agent Creation form
  |
  +-- [Directive List] (accessible from Fleet View)
        |-- All directives with status
        |-- [Enter] on directive -> Directive Detail
        +-- Assignment to ship
```

---

## 6. PRD Use Case Mapping

This section maps every PRD use case to the revised mental model.

### COMM -- Commission and Planning

| PRD UC | PRD Title | Revised Model Mapping |
|--------|----------|----------------------|
| UC-COMM-01 | Create commission from PRD | **JOB C.1**: Create a directive from a PRD. Commission = Directive. |
| UC-COMM-02 | Spawn Ready Room sessions | **JOB D.1**: Initiate directive planning. Sessions spawned for ship's crew (Captain, Commander, Design Officer from the ship's crew roster). |
| UC-COMM-03 | Captain functional analysis | **JOB D.2**: Observe planning progress (Captain's contribution). |
| UC-COMM-04 | Design Officer analysis | **JOB D.2**: Observe planning progress (Design Officer's contribution). |
| UC-COMM-05 | Commander mission decomposition | **JOB D.2**: Observe planning progress (Commander's contribution). |
| UC-COMM-06 | Inter-session message routing | **JOB F.2**: Review communication history (Admiral observes, system routes). |
| UC-COMM-07 | Consensus validation | **JOB D.4**: Review the mission manifest (consensus triggers the review). |
| UC-COMM-08 | Agent to Admiral question | **JOB D.3 / F.1**: Answer agent questions during planning/execution. |
| UC-COMM-09 | Admiral approval gate | **JOB D.4 / D.5 / D.6**: Review manifest, provide feedback, or approve. |
| UC-COMM-10 | Plan persistence | **JOB C.3 / C.4**: Shelve and resume directives. |

### EXEC -- Mission Execution

| PRD UC | PRD Title | Revised Model Mapping |
|--------|----------|----------------------|
| UC-EXEC-01 | Sequence and dispatch missions | **JOB E.1**: Launch a ship on its directive. Commander handles internally. |
| UC-EXEC-02 | Execute RED phase | System internal -- Admiral monitors via **JOB E.2**. |
| UC-EXEC-03 | Execute GREEN phase | System internal -- Admiral monitors via **JOB E.2**. |
| UC-EXEC-04 | Execute REFACTOR phase | System internal -- Admiral monitors via **JOB E.2**. |
| UC-EXEC-05 | Route STANDARD_OPS fast path | System internal -- Admiral monitors via **JOB E.2**. |
| UC-EXEC-06 | Enforce mission termination | System internal -- Admiral notified; may act via **JOB E.4/E.5**. |
| UC-EXEC-07 | Validate demo token | System internal -- Admiral may review via **JOB E.7**. |
| UC-EXEC-08 | Dispatch domain reviewer | System internal -- Admiral monitors via **JOB E.2**. |
| UC-EXEC-09 | Handle review verdict | System internal -- Admiral monitors via **JOB E.2**. |
| UC-EXEC-10 | Report status to Captain | System internal -- Admiral sees via **JOB C.2 / E.3**. |

### GATE -- Verification Gates

| PRD UC | PRD Title | Revised Model Mapping |
|--------|----------|----------------------|
| UC-GATE-01 through 08 | All gate operations | System internal -- deterministic. Admiral sees gate results in event log (**JOB G.3**) and phase tracker (**JOB E.2**). Admiral never directly interacts with gates. |

### HARN -- Harness and Sessions

| PRD UC | PRD Title | Revised Model Mapping |
|--------|----------|----------------------|
| UC-HARN-01 through 06 | Session management | System internal. Admiral configures via **JOB H.2** (role-to-model mapping) and **JOB B.2** (agent model config in roster). |
| UC-HARN-07 | Detect harness availability | **JOB H.3**: Validate harness availability at startup. |
| UC-HARN-08 | Configure harness per role | **JOB H.2** + **JOB B.2**: In revised model, model is configured per-agent in the roster (not just per-role in TOML). This is a meaningful change. |

### STATE -- State Persistence

| PRD UC | PRD Title | Revised Model Mapping |
|--------|----------|----------------------|
| UC-STATE-01 | Initialize Beads | **JOB H.1**: Initialize Ship Commander. |
| UC-STATE-02 through 10 | All state operations | System internal. Admiral benefits from persistence (crash recovery, shelve/resume) but does not interact with Beads directly. |

### TUI -- Terminal User Interface

| PRD UC | PRD Title | Revised Model Mapping |
|--------|----------|----------------------|
| UC-TUI-01 | Planning dashboard | **JOB D.2**: Observe planning progress (Ready Room). |
| UC-TUI-02 | Admiral question modal | **JOB D.3 / F.1**: Answer agent questions. |
| UC-TUI-03 | Plan review overlay | **JOB D.4**: Review mission manifest. |
| UC-TUI-04 | Execution dashboard | **JOB E.2**: Monitor active mission execution (Main Bridge). |
| UC-TUI-05 | Agent status grid | **JOB E.2 / G.2**: Monitor agents, investigate stuck agents. |
| UC-TUI-06 | Live event log | **JOB G.3**: View event log for debugging. |
| UC-TUI-07 | Accept operator commands | **JOB E.4 / E.5 / E.8**: Halt, retry, adjust WIP. |
| UC-TUI-08 | Display system health | **JOB G.1**: Monitor system health. |
| UC-TUI-09 | Wave execution view | **JOB E.2**: Monitor execution (wave progress). |
| UC-TUI-10 | Apply LCARS theme | Cross-cutting -- applies to all screens. |

---

## 7. Gap Analysis

### Concepts in Revised Model NOT in PRD

| Concept | Description | Impact | Recommendation |
|---------|------------|--------|----------------|
| **Fleet** | Collection of ships; portfolio view | Requires new Fleet View screen, fleet-level state, multi-ship navigation | HIGH -- architectural addition. Must decide if fleet is v1 or v2. |
| **Starship as entity** | Named, persistent grouping of agents | Requires Ship data model, ship-to-directive binding, ship commissioning flow | HIGH -- changes the core object model. |
| **Agent Roster** | First-class agent entities with custom config | Requires Agent data model with name/skills/model/prompt, roster management UI, per-agent model validation | HIGH -- significant new feature surface. |
| **Agent personality** | Optional age and backstory | Low impact -- optional fields on agent entity | LOW -- fun feature, add in any phase. |
| **Ship Commissioning** | Ceremony for creating and naming ships | Requires commissioning form, crew assignment flow | MEDIUM -- new workflow. |
| **Per-agent model config** | Model selected per agent, not just per role | Conflicts with PRD's `[roles.captain] model = "opus"` pattern; need to reconcile | MEDIUM -- config model change. |
| **Skill string validation** | Agent skills validated against target harness | Requires skill registry or harness introspection | MEDIUM -- validation complexity. |

### Concepts in PRD NOT explicitly in Revised Model

| Concept | PRD Location | Status in Revised Model |
|---------|-------------|------------------------|
| **Demo Tokens** | UC-EXEC-07 | Implicitly preserved -- part of mission execution. Not called out in revised model but not contradicted. |
| **Surface-Area Locking** | UC-STATE-08 | System internal -- not user-facing. Preserved. |
| **Doctor (Health Monitor)** | Roles table | System internal -- preserved as infrastructure. |
| **Worktree-per-mission** | UC-EXEC-01 | System internal -- preserved. |
| **Protocol Events** | GATE group | System internal -- preserved. |
| **Dual Track (RED_ALERT / STANDARD_OPS)** | Execution Model | System internal -- preserved. Not user-facing in revised model. |
| **Reviewer agent** | Roles table | Could be an Ensign specialist in the revised model. |

### Reconciliation Needed

1. **Per-role vs per-agent model config**: The PRD uses `[roles.captain] model = "opus"` in TOML. The revised model puts the model on the agent entity. These can coexist: TOML provides defaults, agent roster provides overrides.

2. **Ship constraint "one directive at a time"**: The PRD does not model this constraint. It needs explicit enforcement in the state machine (ship.activeDirective is singular).

3. **Fleet View vs Executive Mode**: The PRD's Executive Mode (showing multiple commissions) is conceptually close to the Fleet View but organized around commissions, not ships. The revised model makes ships the organizing unit. Executive Mode should become Fleet View.

4. **Agent lifecycle**: The PRD treats agents as ephemeral (spawned per mission). The revised model treats agents as persistent roster members. Both can coexist: roster agents define configuration, runtime agent instances are ephemeral sessions using that configuration.

---

## 8. Job Priority Assessment

Using Enhanced RICE with Job Criticality.

| Job ID | Job Summary | Reach | Impact | Confidence | Effort | Job Criticality | Score |
|--------|-----------|-------|--------|------------|--------|----------------|-------|
| C.1 | Create directive from PRD | All users | High | High | Low | Required | **Highest** |
| D.1 | Initiate directive planning | All users | High | High | Medium | Required | **Highest** |
| D.4 | Review mission manifest | All users | High | High | Medium | Required | **Highest** |
| D.6 | Approve plan and launch | All users | High | High | Low | Required | **Highest** |
| E.1 | Launch a ship | All users | High | High | Low | Required | **Highest** |
| E.2 | Monitor execution | All users | High | High | High | High | **Highest** |
| D.3 | Answer questions (planning) | All users | High | High | Medium | High (blocking) | **High** |
| E.6 | Answer questions (execution) | All users | High | High | Medium | High (blocking) | **High** |
| E.4 | Halt stuck mission | All users | High | Medium | Low | High (when needed) | **High** |
| B.1 | Create agent (roster) | All users | High | Medium | Medium | Required (revised model) | **High** |
| B.4 | Assign agents to ship | All users | High | Medium | Medium | Required (revised model) | **High** |
| A.1 | Commission a starship | All users | Medium | Medium | Medium | Required (revised model) | **High** |
| A.2 | View fleet status | Multi-ship users | High | Medium | Medium | High | **Medium-High** |
| H.1 | Initialize project | All users | Medium | High | Low | Required (one-time) | **Medium** |
| C.2 | Review directive status | All users | Medium | High | Medium | High | **Medium** |
| D.2 | Observe planning progress | All users | Medium | Medium | Medium | Medium | **Medium** |
| G.1 | Monitor system health | All users | Medium | Medium | Low | Medium | **Medium** |
| E.5 | Retry halted mission | Some users | Medium | Medium | Low | Medium | **Medium** |
| D.5 | Provide feedback | Some users | Medium | Medium | Low | Medium | **Medium** |
| B.2 | Configure agent | All users | Medium | Medium | Medium | Medium | **Medium** |
| E.7 | Review demo tokens | All users | Low | Medium | Low | Medium | **Medium-Low** |
| E.8 | Adjust WIP limits | Some users | Low | Medium | Low | Low | **Low** |
| C.3 | Shelve directive | Some users | Low | Medium | Low | Low | **Low** |
| C.4 | Resume directive | Some users | Low | Medium | Low | Low | **Low** |
| B.3 | Add agent personality | Fun seekers | Low | Low | Low | None | **Lowest** |
| A.4 | Decommission ship | Few users | Low | Low | Low | Low | **Lowest** |
| F.3 | Broadcast message | Few users | Low | Low | Low | Low | **Lowest** |

### Recommended Implementation Order

**Phase 0 (Foundation)**: H.1 (init), B.1 (create agent), A.1 (commission ship), B.4 (assign crew)
**Phase 1 (Planning Loop)**: C.1 (create directive), A.3 (assign directive to ship), D.1 (initiate planning), D.2 (observe planning), D.3 (answer questions), D.4 (review manifest), D.5 (feedback), D.6 (approve)
**Phase 2 (Execution)**: E.1 (launch), E.2 (monitor), E.4 (halt), E.5 (retry), E.6 (answer questions during execution), E.7 (review demo tokens)
**Phase 3 (Fleet and Polish)**: A.2 (fleet view), C.2 (directive status), G.1-G.3 (health/observability), E.8 (WIP), C.3/C.4 (shelve/resume)
**Phase 4 (Delight)**: B.3 (personality), F.3 (broadcast), A.4 (decommission)

---

## Appendix: Job Summary Table

| ID | Category | Job Statement (abbreviated) | Workflow |
|----|----------|---------------------------|----------|
| A.1 | Fleet | Commission a new starship | WF3 |
| A.2 | Fleet | View fleet status | WF10 |
| A.3 | Fleet | Assign directive to ship | WF4 |
| A.4 | Fleet | Decommission/reassign ship | -- |
| B.1 | Roster | Create a new agent | WF2 |
| B.2 | Roster | Configure agent capabilities | WF11 |
| B.3 | Roster | Add personality to agent | WF2 |
| B.4 | Roster | Assign agents to ship crew | WF3 |
| B.5 | Roster | Review agent performance | WF11 |
| C.1 | Directive | Create directive from PRD | WF4 |
| C.2 | Directive | Review directive status | WF6, WF9, WF10 |
| C.3 | Directive | Shelve a directive | WF5 |
| C.4 | Directive | Resume a directive | WF5 |
| D.1 | Planning | Initiate planning | WF5 |
| D.2 | Planning | Observe planning progress | WF5 |
| D.3 | Planning | Answer questions (planning) | WF5, WF7 |
| D.4 | Planning | Review mission manifest | WF5 |
| D.5 | Planning | Provide feedback | WF5 |
| D.6 | Planning | Approve plan | WF5 |
| E.1 | Execution | Launch a ship | WF6 |
| E.2 | Execution | Monitor execution | WF6 |
| E.3 | Execution | Open ship to check progress | WF10 |
| E.4 | Execution | Halt stuck mission | WF8 |
| E.5 | Execution | Retry halted mission | WF8 |
| E.6 | Execution | Answer questions (execution) | WF7 |
| E.7 | Execution | Review demo tokens | WF9 |
| E.8 | Execution | Adjust WIP limits | WF8 |
| F.1 | Comms | Receive/respond to questions | WF5, WF7 |
| F.2 | Comms | Review communication history | WF5 |
| F.3 | Comms | Broadcast to all agents | WF5 |
| G.1 | Health | Monitor system health | WF6 |
| G.2 | Health | Investigate stuck agents | WF8 |
| G.3 | Health | View event log | WF6, WF8 |
| H.1 | Config | Initialize project | WF1 |
| H.2 | Config | Configure role-to-model | WF1 |
| H.3 | Config | Validate harness availability | WF1 |

---

**END OF DOCUMENT**
