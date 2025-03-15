#  Improved precision SQL execution times
## For MySQL and PostgreSQL clients

<div style="border: 2px solid red; padding: 10px; background-color: #ffcccc; color: red;">
  **This project is bleeding-edge, just 2 days old**
</div>

## Background

[Readyset](https://readyset.io) is a next-generation caching product for [MySQL](https://www.mysql.com) and [PostgreSQL](https://postgres.org).
Query response can be so fast that using conventional client tools including `mysql` and `psql` do not provide an accurate performance measurement of individual queries.

Enter **micro-sql**.
A lightweight wrapper written in Go designed to report execution times with improved precision.

It has a few simple features:
* Same interface to run in MySQL or PostgreSQL statements.
* Execute the same query N times included, and provides average execution times across iterations.
* Separate out the query execution and the resultset processing times.
* Limit resultset output to focus on the performance numbers, not the display. Limited content shown to prove results.

### With Caching

Here is a trival example running with the query cached in Readyset executing in <1ms.
NOTE: COUNT(*) is not an ideal query to cache, this is just used to prove a point.

```
$ bin/micro-mysql -u readyset -p -h db -P 3342 imdb
Connected to mysql database 'imdb'!
micro-mysql (18:06:16)> select count(*) from imdb.title;
--------------------------------------------------
count(*)
--------------------------------------------------
11131061
1 rows (0.717209 ms query, 0.756958 ms result)
1 rows (0.330333 ms query, 0.349875 ms result)
1 rows (0.529708 ms query, 0.552375 ms result)
Average: 1 rows (0.525750 ms query, 0.553069 ms result over 3 runs)
--------------------------------------------------
```

### Without Caching

With MySQL, even with the table fully cached in the InnoDB Buffer Pool, execution is consistently 600+ms.

```
bin/micro-mysql -u demouser -p -h db imdb
Connected to mysql database 'imdb'!
micro-mysql (18:04:59)> select count(*) from imdb.title;
--------------------------------------------------
count(*)
--------------------------------------------------
11131061
1 rows (683.124500 ms query, 683.229583 ms result)
1 rows (678.220458 ms query, 678.242208 ms result)
1 rows (680.034666 ms query, 680.281833 ms result)
Average: 1 rows (680.459875 ms query, 680.584541 ms result over 3 runs)
--------------------------------------------------
```

### Larger Resultset

**micro-sql** is not designed for query output. A limited display is provided for summarization. It does process all query results in a single thread to give an indication of a total time for an application.


```
micro-mysql (18:06:20)> select * from name;
--------------------------------------------------
name_id	nconst	name	born	died	updated
--------------------------------------------------
1	nm0000001	Unknown	1899	1987	2025-03-07 22:13:18
2	nm0000002	Unknown	1924	2014	2025-03-07 22:13:18
3	nm0000003	Unknown	1934	<nil>	2025-03-07 22:13:18
4	nm0000004	Unknown	1949	1982	2025-03-07 22:13:18
5	nm0000005	Unknown	1918	2007	2025-03-07 22:13:18
6	nm0000006	Unknown	1915	1982	2025-03-07 22:13:18
7	nm0000007	Unknown	1899	1957	2025-03-07 22:13:18
8	nm0000008	Unknown	1924	2004	2025-03-07 22:13:18
9	nm0000009	Unknown	1925	1984	2025-03-07 22:13:18
10	nm0000010	Unknown	1899	1986	2025-03-07 22:13:18
[...] Output truncated at 10 rows.
14235647 rows (48.136459 ms query, 8389.778333 ms result)
14235647 rows (2.719959 ms query, 8359.327791 ms result)
14235647 rows (2.562458 ms query, 8319.964875 ms result)
Average: 14235647 rows (17.806292 ms query, 8356.357000 ms result over 3 runs)
```

## Command-line Options for micro-sql

Usage:
   bin/micro-mysql <args> <dbname>

- -u \<user>
- -p \<password>
- -h \<host>
- -P \<port> *Optional*
- -c \<count> times to execute query
- -l \<limit> resultset displayed
- \<dbname>

## SQL Commands for micro-sql

- HELP
- EXIT
- SET MICRO COUNT=N
- SET MICRO LIMIT=N
- SELECT 

This program does no parsing of SQL statements, it simple executes the SELECT statement as provided to the respective Go Driver, and reads the result set.
