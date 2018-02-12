# quantum
Simple command line time tracking app

## Usages

### Add
Add a task using `add` or `a`. Arguments are task (mandatory), followed by number of hours (mandatory) and ref (optional).

```
quantum add "Build quantum cli" 10.5 "Ref code: 123"
quantum a "Build quantum cli" 10.5 "Ref code: 123"
```

### List

List the tasks by using `list` or `l`. By default this will show the tasks in the last 7 days. Optional parameter allows you to configure the number of previous days to search over.
```
quantum list
quantum l 10
```
There is also additional support for easy sub commands to search over the past month or year using `list month` or `list year` respectively.
```
quantum list month
quantum list year
```

#### List by task or ref
List tasks by task name or ref using `list task` and `list ref` followed by a space separated list of values to match against
```
quantum list task "QUANTUM-001" "QUANTUM-002"
quantum list ref "QUANTUM-001" "QUANTUM-002"
```

#### List Result
```
+---------+-------+-----------+---------------------+-----------------------------+
|  TASK   | HOURS |    REF    |        DATE         |             UID             |
+---------+-------+-----------+---------------------+-----------------------------+
| FMO-123 |  5.00 |           | 2018-02-12 21:21:17 | 10T0qh4Pm4CPVZw2Z1KaNcjYuXr |
| FMO-456 |  8.00 |           | 2018-02-12 21:21:23 | 10T0rZ93n1Y8kGLTtXU4hOpU24R |
| FMO-789 |  8.00 |           | 2018-02-12 21:21:28 | 10T0s6DZxZ54uWtWcgX6xeP2Vkt |
| FMO-123 |  4.00 |           | 2018-02-12 21:21:33 | 10T0skk088OMfccTTfvb8SeFnK7 |
| FMO-501 |  5.00 | RANDOMREF | 2018-02-12 21:38:37 | 10T2xUQGOJYL9cHown55rc9II3F |
| FMO-110 |  4.50 |           | 2018-02-12 21:49:27 | 10T4H7UzAnUWeVKBUrIQbgCJkBK |
+---------+-------+-----------+---------------------+-----------------------------+
|                                   TOTAL HOURS     |            34 50            |
+---------+-------+-----------+---------------------+-----------------------------+
```

### Delete
Delete a task by using `delete` or `d`. Takes a single uid argument or the record to delete.

```
quantum delete 10Q9i8n6x2djb53ClMSswxFaD9l
quantum d 10Q9i8n6x2djb53ClMSswxFaD9l
```

#### Delete all

Delete all tasks by using quantum 'delete all' or 'd all'.

```
quantum delete all
quantum d all
```
