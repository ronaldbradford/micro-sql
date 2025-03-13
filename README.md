# A Microsecond precision display for MySQL and PostgreSQL clients

<div style="border: 2px solid red; padding: 10px; background-color: #ffcccc; color: red;">
  This project is bleeding-edge, currently only hours old.
</div>

## Background

[Readyset](https://readyset.io) is a next-generation caching product for [MySQL](https://www.mysql.com) and [PostgreSQL](https://postgres.org).
Query response can be so fast that using conventional client tools including `mysql` and `psql` do not provide an accurate performance measurement.

Enter **micro-mysql/micro-psql**, a lightweight wrapper designed to report execution times with higher precision.

It also has a few simple features, like executing the same query N times to get a better sample, limiting results output to focus on the numbers.

```
$ ./micro-mysql -u readyset -p -h picard -c 5 -l 6 -P 3342 imdb
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

```
./micro-mysql -u rbradfor -p -h picard -c 5 -l 6 imdb
Connected to mysql database 'imdb'!
micro-mysql (15:14:18)> SELECT count(*) FROM `imdb`.`title`;
--------------------------------------------------
count(*)
--------------------------------------------------
11114744
1 rows (766.899417 ms query, 767.123250 ms result)
1 rows (744.901541 ms query, 744.910791 ms result)
1 rows (747.766458 ms query, 747.796583 ms result)
1 rows (747.340042 ms query, 747.357083 ms result)
1 rows (746.738208 ms query, 746.760542 ms result)
Average: 1 rows (750.729133 ms query, 750.789650 ms result over 5 runs)
--------------------------------------------------
```
