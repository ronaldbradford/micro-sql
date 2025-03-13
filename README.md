# Microsecond precision execution times
## For MySQL and PostgreSQL clients

<div style="border: 2px solid red; padding: 10px; background-color: #ffcccc; color: red;">
  **This project is bleeding-edge, currently only hours old.***
</div>

## Background

[Readyset](https://readyset.io) is a next-generation caching product for [MySQL](https://www.mysql.com) and [PostgreSQL](https://postgres.org).
Query response can be so fast that using conventional client tools including `mysql` and `psql` do not provide an accurate performance measurement.

Enter **micro-mysql/micro-psql**. A lightweight wrapper written in Go designed to report execution times with higher precision.

It also has a few simple features
* Same interface to run in MySQL or PostgreSQL
* Execute the same query N times, with average execution times.
* Separate query execution and result set processing times.
* Limited resultset output to focus on the performance numbers.


### With Caching

Here is a trival example running with the query cached in Readyset executing in <1ms.

```
$ ./micro-mysql -u readyset -p -h db -c 5 -l 6 -P 3342 imdb
Connected to mysql database 'imdb'!
micro-mysql (15:14:06)> SELECT count(*) FROM `imdb`.`title`;
--------------------------------------------------
count(*)
--------------------------------------------------
11131061
1 rows (1.208625 ms query, 1.288125 ms result)
1 rows (0.853042 ms query, 0.855917 ms result)
1 rows (0.754125 ms query, 0.763833 ms result)
1 rows (0.731625 ms query, 0.736250 ms result)
1 rows (0.767417 ms query, 0.769750 ms result)
Average: 1 rows (0.862967 ms query, 0.882775 ms result over 5 runs)
--------------------------------------------------
```

### Without Caching

With MySQL, even with the table cached in the InnoDB Buffer Pool, execution is consistently 700+ms.

```
./micro-mysql -u rbradfor -p -h db -c 5 -l 6 imdb
Connected to mysql database 'imdb'!
micro-mysql (15:14:18)> SELECT count(*) FROM `imdb`.`title`;
--------------------------------------------------
count(*)
--------------------------------------------------
11131061
1 rows (616.140958 ms query, 616.264500 ms result)
1 rows (611.821541 ms query, 611.838417 ms result)
1 rows (612.376583 ms query, 612.392625 ms result)
1 rows (610.311292 ms query, 610.328750 ms result)
1 rows (617.082791 ms query, 617.100125 ms result)
Average: 1 rows (613.546633 ms query, 613.584883 ms result over 5 runs)
--------------------------------------------------
```

## options

- -u <user>
- -p <password>
- -h <host>
- -P <port> *Optional*
- -c <count> times to execute query
- -l <limit> resultset displayed
- <dbname>

## Commands

- HELP
- EXIT
- SET MICRO COUNT=N
- SET MICRO LIMIT=N
- SELECT 

The command does no parsing of SQL statements, it simple executes the SELECT statement, and reads the resultset.
