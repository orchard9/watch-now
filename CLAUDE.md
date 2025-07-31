# Claude Development Guide

Written by world class golang engineers that pride elegant efficiency and delivering a focused product.

## Memory Management

### Memory Structure
The project maintains five memory files in the `.memory/` directory:

1. **state.md**: Lists every file in the project with a one-line description of its purpose
2. **working-memory.md**: Tracks work progress
   - Recent: 1-3 lines describing completed work
   - Current: 1-3 lines describing work in progress
   - Future: 1-3 lines describing upcoming work
3. **semantic-memory.md**: Contains simple factual statements about the project
4. **vision.md**: Defines the stable long-term vision for the project
5. **tasks.md**: Lists all tasks with 1-3 lines describing what each task is and why it's important

### Task Management
When the user mentions "create tasks", "edit tasks", or "update tasks", they refer to modifying `.memory/tasks.md`. This is the authoritative task list for the project, not any temporary todo tracking systems.

## How to Handle Coding Task Requests

### 0. Gather Requirements
Ask questions until you have 100% confidence in understanding:
- Project specifications
- Requirements
- Expected outcomes
- Edge cases

### 1. Create Project Header
Write a clear header that describes the project's purpose and scope.

### 2. Load Memory and Follow Principles
- First: Load all relevant memory files
- Second: Build elegantly following DRY principles
- Third: Ensure flexibility for future development
- Fourth: Maintain strong adherence to the project vision

### 3. Create 80/20 Tests
When relevant to the task:
- Write simple tests that cover 80% of use cases with 20% effort
- Focus on core functionality first
- Outline expected behavior clearly

### 4. Review Test Coverage
Double-check that tests:
- Follow the 80/20 coverage principle
- Adhere to specifications and vision
- Identify and address any technical debt

### 5. Scaffold Code Structure
Create the file structure with 1-3 line descriptions for each file explaining:
- The file's purpose
- What functionality it provides
- How it fits into the overall architecture

### 6. Validate Architecture
Apply the 80/20 rule to pressure test:
- How code will be called and used
- Whether the architecture supports the requirements
- If changes are needed before implementation

### 7. Implement Code
Write and refine code until:
- All tests pass with "make ci"
- Code meets quality standards
- Implementation matches the design

### 8. User Acceptance Testing
Perform final validation against real-world scenarios:
- Build the binary with `make build`
- Test CLI functionality with `./pg-goer --help` and `./pg-goer --version`
- Test against actual PostgreSQL database if available
- Verify all requirements are met from user's perspective
- Ensure the solution is intuitive and reliable

### 9. Update Memory
After completing the task:
- Update state.md with new files
- Update working-memory.md with completed work
- Add new facts to semantic-memory.md if applicable
- Ensure vision.md remains accurate
- Update `.memory/tasks.md`: mark task as completed, move it to the bottom, and add notes about:
  - How it was implemented
  - What could be improved
  - Any concerns or regressions to watch for

## How to Handle Bug Reports

Written by world class golang engineers that pride elegant efficiency and delivering a focused product.

### 0. Understand the Bug
Ask questions until you have 100% confidence in understanding:
- Exact symptoms and error messages
- Steps to reproduce the issue
- Expected vs actual behavior
- Environmental context (OS, version, data, etc.)

### 1. Reproduce the Failure
- Create a minimal test case that demonstrates the bug
- Write a failing test that captures the exact issue
- Ensure the test fails for the right reasons
- Document the reproduction steps clearly

### 2. Identify Root Cause
- Trace the code path that leads to the failure
- Understand why our existing tests didn't catch this
- Identify the specific logic or assumption that's incorrect
- Consider related areas that might have similar issues

### 3. Fix with Test-First Approach
- Start with the failing test from step 1
- Implement the minimal fix that makes the test pass
- Ensure the fix doesn't break existing functionality
- Follow our coding principles: reliable, elegant, efficient

### 4. Improve Test Coverage
Critical step - prevent regression:
- Add comprehensive tests that would have caught this bug
- Test edge cases and boundary conditions
- Add validation that catches similar issues
- Ensure tests fail meaningfully when broken

### 5. Validate the Fix
- Run `make ci` to ensure all tests pass
- Run `make uat` to validate real-world scenarios
- Test the specific reproduction case from step 0
- Verify related functionality wasn't impacted

### 6. Update Documentation
- Update code comments if the bug revealed unclear logic
- Add examples or warnings for edge cases
- Update user documentation if behavior changed
- Document the fix approach for future reference

### 7. Update Memory
After fixing the bug:
- Update working-memory.md with the bug fix details
- Add lessons learned to semantic-memory.md
- Note any architectural improvements made
- Update `.memory/tasks.md` with prevention measures if needed

## Handling "check gh" Command

When the user says "check gh", perform these actions:

1. **Check GitHub Issues**
   - Use `gh issue list` to see open issues
   - Look for bug reports, feature requests, or questions
   - Add any actionable items to `.memory/tasks.md`

2. **Check GitHub Actions**
   - Use `gh run list` to see recent workflow runs
   - Look for failed builds or tests
   - Investigate any failures and add fixes to `.memory/tasks.md`

3. **Update Task List**
   - Add new tasks for any issues found to `.memory/tasks.md`
   - Prioritize based on severity (build failures = high priority)
   - Include issue/run numbers in task descriptions for tracking%
