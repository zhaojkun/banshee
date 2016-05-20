Banshee
=======

Banshee is a real-time anomalies(outliers) detection system for periodic
metrics.

[![Build Status](https://travis-ci.org/eleme/banshee.svg?branch=master)](https://travis-ci.org/eleme/banshee)
[![GoDoc](https://godoc.org/github.com/eleme/banshee?status.svg)](https://godoc.org/github.com/eleme/banshee)
[![Join the chat at https://gitter.im/eleme/banshee](https://badges.gitter.im/eleme/banshee.svg)](https://gitter.im/eleme/banshee?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

![snap-01](snap/01.png)

Case
----

For example, a website api's response time is reported to banshee from statsd
every 10 seconds:

```
20, 21, 21, 22, 23, 19, 18, 21, 22, 20, ..., 300
```

The latest `300` will be catched.

Features
--------

* Designed for periodic metrics.
* Dynamic threshold analyzation via 3-sigma.
* Also supports fixed-threshold alert option.
* Provides an alert rule management panel.
* No extra storage services required.

Requirements
------------

1. Go >= 1.5.
2. Node and gulp.

Build
-----

1. Clone this repo.
2. Build binary via `make`.
3. Build static files via `make static`.

Usage
-----

```bash
$ ./banshee -c filename
```

Example configuration file is [config/exampleConfig.yaml](config/exampleConfig.yaml).

Statsd Integration
------------------

1. Install [statsd-banshee](https://www.npmjs.com/package/statsd-banshee) to forward
   metrics to banshee.

   ```bash
   $ cd path/to/statsd
   $ npm install statsd-banshee
   ```

2. Add `statsd-banshee` to statsd backends in config.js:

   ```js
	{
	, backends: ['statsd-banshee']
	, bansheeHost: 'localhost'
	, bansheePort: 2015
	}
   ```

Deployment
----------

Banshee is a single-host program, its detection is fast enough in our case,
we don't have a plan to expand it now.

We are using a Python script ([deploy.py](deploy.py) via [fabric](http://www.fabfile.org/))
to deploy it to remote host:

```
python deploy.py -u hit9 -H remote-host:22 --remote-path "/service/banshee"
```

Upgrade
-------

Just pull the latest [tag release](https://github.com/eleme/banshee/releases).
*Please don't use master branch directly, checkout to a tag instead.*

Generally we won't release not-backward-compatiable versions, if any, related notes
would be added to the [changelog](changelog).

Alert Command
-------------

Banshee requires a command, normally a script to send alert messages.

It should be called from command line like this:

```bash
$ ./alert-command <JSON-String>
```

The JSON string example can be found at [alerter/exampleCommand/echo.go](alerter/exampleCommand/echo.go).

Philosophy
----------

But how do you really analyze the anomalous metrics? Via 3-sigma:

```python
>>> import numpy as np
>>> x = np.array([40, 52, 63, 44, 54, 43, 67, 54, 49, 45, 48, 54, 57, 43, 58])
>>> mean = np.mean(x)
>>> std = np.std(x)
>>> (80 - mean) / (3 * std)
1.2608052883472445 # anomaly, too big
>>> (20 - mean) / (3 * std)
-1.3842407711224991 # anomaly, too small
```

For further implementation introduction, please checkout [intro.md](intro.md).

Authors
-------

Thanks to our [contributors](https://github.com/eleme/banshee/graphs/contributors).

License
-------

MIT Copyright (c) 2015 - 2016 Eleme, Inc.
