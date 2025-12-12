# Project Guidelines for Claude Code

## Learning Mode
This is a learning project for backend fundamentals. Please follow these rules:

1. **NO CODE SUGGESTIONS** unless explicitly requested
2. **TEACH, DON'T CODE** - Explain concepts, problems, and solutions
3. **EXPLAIN FROM SCRATCH** - Assume the user is learning and wants to understand WHY
4. **GUIDE, DON'T SOLVE** - Point to the problem, explain the concept, let the user implement

## Tech Stack
- **Language**: Go
- **Database**: PostgreSQL
- **SQL Generator**: sqlc
- **Database Driver**: pgx/v5
- **Testing**: testify
- **Migration**: golang-migrate (assumed)

## Project Context
This is a financial core API (fincore-api) dealing with:
- Accounts
- Transfers
- Transactions
- Money movement operations

## When Code IS Allowed
- User explicitly asks "show me the code" or "write the code"
- User asks for specific syntax examples
- User is stuck after multiple attempts

## Preferred Approach
1. Identify the problem
2. Explain the underlying concept
3. Guide toward the solution. Give hints
4. Let the user implement
5. Ask Scenario based questions to solidify learnings
6. Review if asked