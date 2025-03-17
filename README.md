#  Improved precision SQL execution times
## For MySQL and PostgreSQL clients

<div style="border: 2px solid red; padding: 10px; background-color: #ffcccc; color: red;">
  **This project is bleeding-edge, just a few days old**
</div>

## Background

[Readyset](https://readyset.io) is a next-generation caching product for [MySQL](https://www.mysql.com) and [PostgreSQL](https://postgres.org).
Measuring auery response can be difficult with conventional client tools including `mysql` and `psql` to provide an accurate precision measurement of individual queries.

Enter **micro-sql**.

A lightweight wrapper written in Go designed to report execution times with improved precision.

This has a few simple features to assist with instrumenting:
* Same interface to run SQL statements in MySQL or PostgreSQL.
* Execute the same query 'N' times built-in, providing average execution times across 'N' iterations.
* Separate the actual query execution and the resultset processing times.
* Limit results output to focus on the performance numbers, not the data display. Limited content is shown to prove results are produced from the queries.

### Example Using Readyset Caching

Here is a simple example running the query cached in Readyset executing in <1ms, close to 0.5ms.

```
$ bin/micro-mysql -u readyset -p *** -h db -P 3342 imdb
micro-sql version: v0.5.3-7b80f3aa18
Connected to mysql database 'imdb'!
micro-mysql (13:01:07)> select count(*) from imdb.title;
------------------------------------------------------------
count(*)
------------------------------------------------------------
11131061
1 rows (0.551 ms query, 0.023 ms result)
1 rows (0.485 ms query, 0.010 ms result)
1 rows (0.504 ms query, 0.010 ms result)
Average: 1 rows (0.513 ms query, 0.014 ms result, 3 executions)
------------------------------------------------------------
```

NOTE: COUNT(*) is not an ideal query to cache, this is just to demonstrate a simple SQL statement.

### Example Without Caching

With MySQL, even with the table fully cached in the InnoDB Buffer Pool, execution is consistently 600+ms.

```
bin/micro-mysql -u $USER -p *** -h db -c 4 imdb
micro-sql version: v0.5.3-7b80f3aa18
Connected to mysql database 'imdb'!
micro-mysql (13:00:43)> select count(*) from imdb.title;
------------------------------------------------------------
count(*)
------------------------------------------------------------
11131061
1 rows (618.272 ms query, 0.030 ms result)
1 rows (617.303 ms query, 0.014 ms result)
1 rows (617.224 ms query, 0.015 ms result)
1 rows (616.721 ms query, 0.015 ms result)
Average: 1 rows (617.380 ms query, 0.019 ms result, 4 executions)
------------------------------------------------------------
```

### Output with larger Resultsets

**micro-sql** is not designed for query output. A limited display is provided for verification. It does process all the query results in a single thread to give an indication of a total time for an application.

In this example, the resultset has 14M rows. While the query is executed in ~1ms, the full resultset takes another 7000+ms (i.e. 7 seconds) for the client to receive the data. If your application then wanted to render some of this, that would take longer.


```
micro-mysql (13:01:07)> select * from name;
------------------------------------------------------------
name_id	nconst	name	born	died	updated
------------------------------------------------------------
1	nm0000001	Fred Astaire	1899	1987	2025-03-13 18:47:56
2	nm0000002	Lauren Bacall	1924	2014	2025-03-13 18:47:56
3	nm0000003	Brigitte Bardot	1934	<nil>	2025-03-13 18:47:56
4	nm0000004	John Belushi	1949	1982	2025-03-13 18:47:56
5	nm0000005	Ingmar Bergman	1918	2007	2025-03-13 18:47:56
6	nm0000006	Ingrid Bergman	1915	1982	2025-03-13 18:47:56
7	nm0000007	Humphrey Bogart	1899	1957	2025-03-13 18:47:56
8	nm0000008	Marlon Brando	1924	2004	2025-03-13 18:47:56
9	nm0000009	Richard Burton	1925	1984	2025-03-13 18:47:56
10	nm0000010	James Cagney	1899	1986	2025-03-13 18:47:56
... ... Output truncated at 10 rows.
14235647 rows (1.795 ms query, 7193.284 ms result)
14235647 rows (1.072 ms query, 7192.926 ms result)
14235647 rows (1.149 ms query, 7193.024 ms result)
Average: 14235647 rows (1.339 ms query, 7193.078 ms result, 3 executions)
------------------------------------------------------------
```

## Command-line Options

Usage:
```
  $ bin/micro-mysql <args> <dbname>
```

Where args include:

### Required

- `-u <user>`
- `-p <password>`
- `-h <host>`

### Optional

- `-P <port>`
- `-c <count>` times to execute query
- `-l <limit>` rows displayed

## Supported SQL Commands

- `SELECT`
- `SHOW`
- `EXIT`
- `SET MICRO COUNT=N`
- `SET MICRO LIMIT=N`
- `HELP`

This program does no parsing of SQL statements, it simple executes the SELECT|SHOW statement as provided to the respective Go Driver, and reads the result set.
