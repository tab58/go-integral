# go-integral

A database seed script generator based on a PostgreSQL schema.

It creates structs based on the SQL schema and some helper functions to add them into the database. It also calculates the dependency graph of the entities and inserts them in an order that doesn't cause problems with the foreign key constraints.
