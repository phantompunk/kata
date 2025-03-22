# Kata CLI: tdd tool for LeetCode

Kata is a command-line client for LeetCode with the goal using TDD to solve problems.  

```bash
❯ kata download --problem 3sum
❯ kata test --problem 3sum
❯ kata list
╔═════╦══════════════════════════════╦════════════╦════╦════════╗
║ ID  ║ Name                         ║ Difficulty ║ Go ║ Python ║
╠═════╬══════════════════════════════╬════════════╬════╬════════╣
║ 15  ║ 3Sum                         ║ Medium     ║ ✅ ║ ❌     ║
║ 18  ║ 4Sum                         ║ Medium     ║ ✅ ║ ❌     ║
║ 128 ║ Longest Consecutive Sequence ║ Medium     ║ ✅ ║ ❌     ║
╚═════╩══════════════════════════════╩════════════╩════╩════════╝
```

## ✨ Features 
- Solve problems your way with your local environment
- Stubs solutions, test files, and readmes locally 
- Saves problems locally for offline practice
- TODO - Submit solutions directly to LeetCode
- TODO - Authenticate automattically using browswer cookies
- TODO - Generate tests with edge cases
- TODO - Check the complexity of your solution
- Built for Go and Python practice - Other languages supported soon




## ⚡️ Quick start


Download and install Go. Version 1.23.0 (or higher) is required.

Then, use the go install command:

`go install github.com/phantompunk/kata@latest`

### Examples
```bash
kata download --problem 15									 # stub problem using question id
kata download --problem 3sum                 # stub problem using question slug

kata download -p 3sum --language go          # stub problem, specifying Go language
kata download -p 3sum -l go --open           # stub problem, open using $EDITOR

kata test --problem 3sum --language go       # test solution against leetcode servers

kata list                                    # list completed problems
kata list --markdown                         # list completed problems as Markdown
kata quiz                                    # select a random question to practice

kata settings                                # open config settings using $EDITOR
kata login                                   # use session and token from browser cookies
kata settings token=${LEETCODE_TOKEN}        # set LeetCode token to enable submissions
kata settings session=${LEETCODE_SESSION}    # set LeetCode session to enable submissions
```
