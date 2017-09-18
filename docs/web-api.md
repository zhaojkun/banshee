Web HTTP API
============

Error
-----

An error has a code and msg, example

```
POST /api/project -d {"name": "duplicated"}

403
{
  "code": 403,
  "msg": "Duplicate project name"
}
```

Auth
----

Some apis may require basic auth, if the auth fails, `Unauthorized`
would be returned:

```
401
Unauthorized
```

Config
------

1. Get config (basic auth required):

   ```
   GET /api/config

   200
   {
     "interval": 10,
     "period": 86400,
     "expiration": 604800,
     ...
   }
   ```

2. Get interval:

   ```
   GET /api/interval

   200
   {
     "interval": 10,
   }
   ```

Team
-------

1. Get all teams.

    ```
    GET /api/teams
    
    200
    [
        {
            "id": 1,
            "name": "Default",
            "numProjects": 0
        },
        ...
    ]
    ```
2. Create team (basic auth required).

   ```
   POST /api/team -d {"name": "myNewTeam"}   
   
   200
   {
       "id": 2,
       "name": "myNewTeam"
   }
   ```
3. Get team by id.

    ```
    GET /api/team/:id
    
    200
    {
        "id": 1,
        "name": "Default"
    }
    ```
4. Update team by id (basic auth required).

    ```
    PATCH /api/team/:id -d {"name": "newName"}
    
    200
    {
        "id": 1,
        "name": "newName"
    }
    ```
5. Delete team by id (basic auth required).
   
   ```
   DELETE /api/team/:id
   
   200
   ```

Project
-------

1. Get projects by team id.

    ```
    GET /api/team/:id/projects
    
    200
    [
        {
            "id": 1,
            "name": "test",
            "numRules": 2,
        },
        ...
    ]
     ```
     
2. Create project by team id (basic auth required).

   ```
   POST /api/team/:id/project -d {"name": "myNewProject"}

   200
   {
     "id": 2,
     "name": "foo",
     "enableSilent": false,
     "silentTimeEnd": 0,
     "silentTimeStart": 0,
     "teamID": 1
   }
   ```

3. Get project by id.

   ```
   GET /api/project/:id

   200
   {
     "enableSilent": false,
     "id": 1,
     "name": "test",
     "silentTimeEnd": 0,
     "silentTimeStart": 0,
     "teamID": 1
   }
   ```

4. Update project (basic auth required).

   ```
   PATCH /api/project/:id -d {"name": "newName","teamID":1,"silentTimeStart":1,"silentTimeEnd":10}

   200
   {
     "enableSilent": false,
     "id": 1,
     "name": "newName",
     "silentTimeEnd": 10,
     "silentTimeStart": 1,
     "teamID": 1
   }
   ```

5. Delete project by id (basic auth required).

   ```
   DELETE /api/project/:id

   200
   ```

6. Get projects by user id (basic auth required).

   ```
   GET /api/user/:id/projects

   200
   [
     {
       "enableSilent": true,
       "id": 1,
       "name": "test",
       "silentTimeEnd": 0,
       "silentTimeStart": 0,
       "teamID": 1
     },
     ...
   ]
   ```

7. Get all projects 
   
   ```
   GET /api/projects
   
   200
   [
       {
           "id": 1,
           "name": "test",
           "numRules": 2,
           "teamID": 1
       },
       ...
   ]
   ```
   
User
----

1. Get all users (basic auth required).

   ```
   GET /api/users

   200
   [
     {
       "email": "xiaoming@gmail.com",
       "enableEmail": false,
       "enablePhone": true,
       "id": 2,
       "name": "xiaoming",
       "phone": "18718718718",
       "ruleLevel": 0,
       "universal": true
     },
     ...
   ]
   ```

2. Create user (basic auth required).

   ```
   POST /api/user -d {
     "name": "jack",
     "email": "jack@gmail.com",
     "enableEmail": false,
     "phone": "18718718718",
     "enablePhone": true,
     "universal": true
   }

   200
   {
     "id": 1,
     "name": "jack",
     "email": "jack@gmail.com",
     "enableEmail": false,
     "phone": "18718718718",
     "enablePhone": true,
     "universal": true,
     "ruleLevel": 0
   }
   ```

3. Get user by id (basic auth required).

   ```
   GET /api/user/:id

   200
   {
     "email": "jack@gmail.com",
     "enableEmail": false,
     "enablePhone": true,
     "id": 1,
     "name": "jack",
     "phone": "18718718718",
     "ruleLevel": 0,
     "universal": true
   }
   ```

4. Update user by id (basic auth required).

   ```
   PATCH /api/user/:id -d {
     "name": "jack",
     "email": "jack@gmail.com",
     "enableEmail": true,
     "phone": "18718718718",
     "enablePhone": true,
     "universal": true
   }

   200
   {
     "email": "jack@gmail.com",
     "enableEmail": true,
     "enablePhone": true,
     "id": 1,
     "name": "jack",
     "phone": "18718718718",
     "ruleLevel": 0,
     "universal": true
   }
   ```

5. Delete user by id (basic auth required).

   ```
   DELETE /api/user/:id

   200
   ```
6. Add user to project by username (basic auth required).
   
   ```
   POST /api/project/:id/user -d { name :"admin" }
   
   200
   ```


7. Get users by project id (basic auth required).

   ```
   GET /api/project/:id/users

   200
   [
     {
       "email": "xiaoming@gmail.com",
       "enableEmail": false,
       "enablePhone": true,
       "id": 2,
       "name": "xiaoming",
       "phone": "18718718718",
       "ruleLevel": 0,
       "universal": true
     },
     ...
   ]
   ```

8. Copy rules bewteen users based on userID (basic auth required).
   ```
   POST /api/users/copy -d {"from": 1 ,"to": 2 }
   
   200
   ```

Rule
----

1. Get rules by project id (basic auth required).

   ```
   GET /api/project/:id/rules

   200
   [
     {
       "comment": "测试数据",
        "disabled": false,
        "disabledAt": "2016-05-23T15:51:57.78674871+08:00",
        "disabledFor": 0,
        "id": 1,
        "level": 0,
        "neverFillZero": false,
        "numMetrics": 1,
        "pattern": "timer.mean_90.*",
        "projectID": 1,
        "thresholdMax": 0,
        "thresholdMin": 0,
        "trendDown": false,
        "trendUp": true
     },
     {
        "comment": "",
        "disabled": false,
        "disabledAt": "2016-05-20T16:17:59.535476749+08:00",
        "disabledFor": 0,
        "id": 2,
        "level": 0,
        "neverFillZero": false,
        "numMetrics": 1,
        "pattern": "timer.count_ps.*",
        "projectID": 1,
        "thresholdMax": 0,
        "thresholdMin": 0,
        "trendDown": true,
        "trendUp": true
     }
   ]
   ```

2. Create a rule for given project (basic auth required).

   ```
   POST /api/project/:id/rule -d {
     "pattern": "timer.count_ps.foo",
     "trendUp": true,
     "trendDown": false,
     "thresholdMax": 100,
     "thresholdMin": 0,
     "comment": "interface foo",
     "level": 1,
     "disabled": false,
     "disabledFor": 0,
     "neverFillZero": false
   }

   200
   {
     "id": 4,
     "projectID": 1,
     "pattern": "timer.count_ps.foo",
     "trendUp": true,
     "trendDown": false,
     "thresholdMax": 100,
     "thresholdMin": 0,
     "numMetrics": 0,
     "comment": "interface foo",
     "level": 1,
     "disabled": false,
     "disabledFor": 0,
     "disabledAt": "2016-05-23T18:12:36.829057386+08:00",
     "neverFillZero": false
   }
   ```

3. Update a rule by id (basic auth required).

   ```
   PATCH /api/rule/:id -d {
     "pattern": "timer.mean_90.bar",
     "trendUp": true,
     "trendDown": false,
     "thresholdMax": 40,
     "thresholdMin": 0,
     "comment": "interface foo",
     "level": 1,
     "disabled": false,
     "disabledFor": 0,
     "neverFillZero": false
   }

   200
   {
     "id": 4,
     "projectID": 1,
     "pattern": "timer.count_ps.foo",
     "trendUp": true,
     "trendDown": false,
     "thresholdMax": 100,
     "thresholdMin": 0,
     "numMetrics": 0,
     "comment": "interface foo",
     "level": 1,
     "disabled": false,
     "disabledFor": 0,
     "disabledAt": "2016-05-23T18:12:36.829057386+08:00",
     "neverFillZero": false
   }
   ```

4. Delete rule by id (basic auth required).

   ```
   DELETE /api/rule/:id

   200
   ```

Metric
------

1. Get indexes.

   ```
   GET /api/metric/indexes?limit=<number>&sort=<up|down>&pattern=<string>
   OR
   GET /api/metric/indexes?limit=<number>&sort=<up|down>&project=<id>

   200
   [
     {
       "average": 0.9977011494252875,
       "link": 3,
       "matchedRules": [
         {
           "comment": "",
           "disabled": false,
           "disabledAt": "0001-01-01T00:00:00Z",
           "disabledFor": 0,
           "id": 2,
           "level": 0,
           "neverFillZero": false,
           "numMetrics": 0,
           "pattern": "timer.count_ps.*",
           "projectID": 1,
           "thresholdMax": 0,
           "thresholdMin": 0,
           "trendDown": true,
           "trendUp": true
         }
       ],
       "name": "timer.count_ps.bar",
       "score": 0.03221414267914575,
       "stamp": 1463999915
     },
     ...
   ]
   ```

2. Get metric values.

   ```
   GET /api/metric/data?start=<timestamp>&stop=<timestamp>&name=<string>

   200
   [
     {
       "name": "timer.count_ps.bar",
       "stamp": 1464057219,
       "value": 1,
       "score": 0.06337242505244704,
       "average": 0.9965116279069768,
       "link": 3
     },
     ...
   ]
   ```

3. Get metric matched rules.

   ```
   GET /api/metric/rules/:name
   
   200
   [
     {
       "id": 2,
       "projectID": 1,
       "pattern": "timer.count_ps.*",
       "trendUp": true,
       "trendDown": true,
       "thresholdMax": 0,
       "thresholdMin": 0,
       "numMetrics": 0,
       "comment": "",
       "level": 0,
       "disabled": false,
       "disabledFor": 0,
       "disabledAt": "0001-01-01T00:00:00Z",
       "neverFillZero": false
     },
     ...
   ]
   ```

Event
-----

1. Get events by project id. (default `past`: `3600*24`), baisc auth required:

   ```
   GET /api/project/:id/events?past=<number>

   200
   [
     {
       "id": "f2d44be1b10dd44875f86a9c1eacfac912ecc45e",
       "ruleID": 10,
       "projectID": 12,
       "level": 2,
       "comment": "Add note",
       "name": "timer.count_ps.note.add",
       "stamp": 1464166573,
       "score": 1.7112404844428821,
       "average": 1.2682274247491634,
       "value": 21.4
     },
     ...
   ]
   ```
2. Get events by time range. (default `past`: `3600*24`, default `end`: `now`):

   ```
   GET /api/events?end=<number>&past=<number>
   200
   [
     {
       "id": "f2d44be1b10dd44875f86a9c1eacfac912ecc45e",
       "ruleID": 10,
       "projectID": 12,
       "level": 2,
       "comment": "Add note",
       "name": "timer.count_ps.note.add",
       "stamp": 1464166573,
       "score": 1.7112404844428821,
       "average": 1.2682274247491634,
       "value": 21.4
     },
     ...
   ]
   ```

WebHook
-----

1. When there is an event, the following payload will be sent to given URL:

  ```
  HTTP Method: POST
  Content-Type: application/json
  PayLoad :
  {
    "id": "ab372c88c08ada73bd35feea740896d340de770e",
    "comment": "add note",
    "timestamp": 1505149763562,
    "text": "{低等级 20:39:00 test 4 efe0f31} test大于设定阈值 当前值1.11,阈值1",
    "grafanaLink": "grafanaLink",
    "ruleLink": "ruleLink",
    "metric": {
      "name": "timer.count_ps.note.add",
      "stamp": 1477361415,
      "value": 1111,
      "score": 0,
      "average": 1111,
      "link": 1
    },
    "rule": {
      "id": 155,
      "projectID": 14,
      "pattern": "timer.count_ps.note.add",
      "trendUp": false,
      "trendDown": false,
      "thresholdMax": 30,
      "thresholdMin": 0,
      "numMetrics": 1,
      "comment": "add note",
      "level": 2,
      "disabled": false,
      "disabledFor": 0,
      "disabledAt": "2016-11-02T11:00:10.611135452+08:00",
      "trackIdle": false,
      "neverFillZero": false
      },
    "project": {
      "id": 14,
      "name": "note",
      "enableSilent": true,
      "silentTimeStart": 0,
      "silentTimeEnd": 6,
      "teamID": 4
      },
    "team": {
      "id": 4,
      "name": "note"
     },
    "alarmUsers": [
               {
                 "email": "xiaoming@gmail.com",
                 "enableEmail": false,
                 "enablePhone": true,
                 "id": 2,
                 "name": "xiaoming",
                 "phone": "18718718718",
                 "ruleLevel": 2,
                 "universal": true
               },
               ...
    ]
    }
  ```

Misc
----

1. Get banshee version.

   ```
   GET /api/version

   200
   {
     "version": "0.2.2"
   }
   ```

2. Get webapp default language.

   ```
   GET /api/language
   
   200
   {
     "language": "zh"
   }
   ```

3. Get health information.

   ```
   GET /api/info
   
   200
   {
     "aggregationInterval": 60,
     "numIndexTotal": 3,
     "numClients": 1,
     "numRules": 3,
     "detectionCost": 1.0395784166666664,
     "filterCost": 0.01339311111111111,
     "queryCost": 0.9379515,
     "numMetricIncomed": 18,
     "numMetricDetected": 12,
     "numAlertingEvents": 0
   }
   ```
