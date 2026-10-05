# toylang

A small programming language built from scratch in Go. No parser generators, no VM libraries. Just: lexer → parser → compiler → bytecode → VM.

## Features

- **Full compiler pipeline**: lexer, recursive-descent parser, bytecode compiler, stack-based VM
- **Language features**:
  - Variables (`let x = 5`)
  - Arithmetic (`+ - * / %`)
  - Comparison (`== != < > <= >=`)
  - Logic (`and or not`)
  - Conditionals (`if / else if / else`)
  - Loops (`while`)
  - Functions (`fn add(a, b) { return a + b }`)
  - Strings and numbers
  - Comments (`# this is a comment`)
- **CLI**: `run`, `parse`, `compile` for debugging

## Not yet supported

- **Recursion** — currently fails with a clean error (`stack frame handling is planned for v0.2`)
- Arrays, dictionaries, classes
- Closures / first-class functions

## Installation

```bash
git clone https://github.com/arjunkapoorkapadiya-rgb/toylang.git
cd toylang
go build -o toylang.exe