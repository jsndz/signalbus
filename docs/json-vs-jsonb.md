# PostgreSQL: `json` vs `jsonb`

PostgreSQL provides two JSON data types: **`json`** and **`jsonb`**.  


stored as:
plaintext 

parsed binary
Ignores whitespace and key order and no duplicates

Perfomance:
slow query

fast since parsed and stored in a structured format


application:
cant change data need exactly

lookup inside json
