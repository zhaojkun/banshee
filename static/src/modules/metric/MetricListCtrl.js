/*@ngInject*/
module.exports = function($scope, $rootScope, $timeout, $stateParams,
                          $translate, Metric, Config, Project, Util) {
  var chart = require('./chart');
  var cubism;
  var initOpt;
  var isInit = false;

  $rootScope.currentMetric = true;
  $scope.projectId = $stateParams.project;
  $scope.past = null;
  $scope.pastUsed = false;

                            
  $scope.dateTimes = [
    {label: 'METRIC_PAST_NOW', seconds: 0},
    {label: 'METRIC_PAST_3HOURS_AGO', seconds: 3 * 3600},
    {label: 'METRIC_PAST_6HOURS_AGO', seconds: 6 * 3600},
    {label: 'METRIC_PAST_1DAY_AGO', seconds: 24 * 3600},
    {label: 'METRIC_PAST_2DAYS_AGO', seconds: 48 * 3600},
    {label: 'METRIC_PAST_3DAYS_AGO', seconds: 3 * 24 * 3600},
    {label: 'METRIC_PAST_4DAYS_AGO', seconds: 4 * 24 * 3600},
    {label: 'METRIC_PAST_5DAYS_AGO', seconds: 5 * 24 * 3600},
    {label: 'METRIC_PAST_6DAYS_AGO', seconds: 6 * 24 * 3600},
    {label: 'METRIC_PAST_7DAYS_AGO', seconds: 7 * 24 * 3600}
  ];

  $scope.limitList = [
    {label: 'METRIC_LIMIT_1', val: 1},
    {label: 'METRIC_LIMIT_30', val: 30},
    {label: 'METRIC_LIMIT_50', val: 50},
    {label: 'METRIC_LIMIT_100', val: 100},
    {label: 'METRIC_LIMIT_500', val: 500},
    {label: 'METRIC_LIMIT_1000', val: 1000}
  ];

  $scope.sortList = [
    {label: 'METRIC_TREND_UP', val: 'up'},
    {label: 'METRIC_TREND_DOWN', val: 'down'}
  ];

  $scope.typeList = [
    {label: 'METRIC_TYPE_VALUE', val: 'v'},
    {label: 'METRIC_TYPE_SCORE', val: 'm'}
  ];

  $scope.autoComplete = {searchText: ''};

  initOpt = {
    project: $stateParams.project,
    pattern: $stateParams.pattern,
    datetime: $scope.dateTimes[0].seconds,
    limit: $scope.limitList[2].val,
    sort: $scope.sortList[0].val,
    type: $scope.typeList[0].val,
    status: false
  };

  if (typeof $stateParams.past !== 'undefined') {
    $scope.past =
        Util.secondsToTimespanString(Util.timeSpanToSeconds($stateParams.past));
  }

  $scope.filter = angular.copy(initOpt);

  $scope.toggleCubism = function() {
    $scope.filter.status = !$scope.filter.status;
    if (!$scope.filter.status) {
      buildCubism();
    } else {
      cubism.stop();
    }
  };

  $scope.restart = function() {
    $scope.filter = angular.copy(initOpt);

    if ($scope.initProject) {
      $scope.project = $scope.initProject;
      $scope.autoComplete.searchText = $scope.project.name;
    } else {
      $scope.project = '';
      $scope.autoComplete.searchText = '';
    }

    $scope.spinner = true;
    $timeout(function() { $scope.spinner = false; }, 1000);

    buildCubism();
  };

  $scope.searchPattern = function() {
    $scope.filter.project = '';
    $scope.autoComplete.searchText = '';
    buildCubism();
  };

  $scope.searchProject = function(project) {
    $scope.filter.project = project.id;
    $scope.filter.pattern = '';
    $scope.project = project;
    $scope.projectId = project.id;

    buildCubism();
  };

  $scope.$on('$destroy', function() { $rootScope.currentMetric = false; });

  /**
   * watch filter.
   */
  function watchAll() {
    $scope.$watchGroup(
      ['filter.datetime', 'filter.limit', 'filter.sort', 'filter.type'],
      function() { buildCubism(); });
  }


  function loadData() {
    Project.getAllProjects().$promise.then(function(res) {
      var projectId = parseInt($stateParams.project);
      $scope.projects = res;
      var teams = {};
      for(var i in $scope.projects){
        var project = $scope.projects[i];
        teams[project.id] = project.teamID;
      }
      $scope.teams = teams;
      if (projectId) {
        $scope.projects.forEach(function(el) {
          if (el.id === projectId) {
            $scope.autoComplete.searchText = el.name;
            $scope.initProject = el;
            $scope.project = el;
            $scope.teamID = el.teamID;
          }
        });
      }
    });

    Config.getInterval().$promise.then(function(res) {
      $scope.filter.interval = res.interval;

      setIntervalAndRunNow(buildCubism, 10 * 60 * 1000);

      watchAll();
    });

    Config.getGraphiteUrl().$promise.then(
        function(res) { $scope.graphiteUrl = res.graphiteUrl; });
  }

  function buildCubism() {
    var params = {
      limit: $scope.filter.limit,
      sort: $scope.filter.sort,
    };
    if ($scope.filter.project) {
      params.project = $scope.filter.project;
    } else {
      params.pattern = $scope.filter.pattern;
    }

    if ($scope.past && !$scope.pastUsed) {
      $scope.filter.datetime = Util.timeSpanToSeconds($scope.past);
      $scope.pastUsed = true;
    }

    chart.remove();

    isInit = false;

    cubism = chart.init({
      selector: '#chart',
      serverDelay: $scope.filter.datetime * 1000,
      step: $scope.filter.interval * 1000,
      stop: false,
      type: $scope.filter.type
    });

    Metric.getMetricIndexes(params).$promise.then(function(res) { plot(res); });
  }

  /**
   * Plot.
   */
  function plot(data) {
    var name, i, metrics = [];
    for (i = 0; i < data.length; i++) {
      name = data[i].name;
      metrics.push(feed(name, data, refreshTitle));
    }

    return chart.plot(metrics);
  }
  function refreshTitle(data) {
    var _titles = d3.selectAll('.title')[0];
    if (isInit) {
      return;
    }
    _titles.forEach(function(el, index) {
      var _el = _titles[index];
      var currentEl = data[index];
      var className = getClassNameByTrend(currentEl.score, currentEl.stamp);
      var str;
      var _box = ['<div class="box"><span>' +
                  $translate.instant('METRIC_METRIC_RULES_TEXT') +
                  '<span class="icon-tr"></span></span><ul>'];

      for (var i = 0; i < currentEl.matchedRules.length; i++) {
        var rule = currentEl.matchedRules[i];
        var teamID = $scope.teams===null?-1:$scope.teams[rule.projectID];
        _box.push('<li><a href="#/admin/team/'+teamID+'/project/' + rule.projectID + '?rule=' +
                  rule.id + '">' + rule.pattern + '</a></li>');
      }
      _box.push('</ul></div>');

      // Graphite name.
      var graphiteUrlHtml = '';
      if ($scope.graphiteUrl || $scope.graphiteUrl.length > 0) {
        var graphiteName = Util.getGraphiteName(currentEl.name);
        var graphiteUrl = Util.format($scope.graphiteUrl, graphiteName);
        graphiteUrlHtml = Util.format(
            '<a class="graphite-link" href="%s" target="_blank">%s</a>',
            graphiteUrl, $translate.instant('METRIC_CHART_TEXT'));
      }

      str = [
        '<a href="#/metric?pattern=' + currentEl.name + '" class="' + className +
            '">',
        getTextByTrend(currentEl.score),
        currentEl.name,
        '</a>',
        graphiteUrlHtml,
        _box.join(''),
      ].join('');

      _el.innerHTML = str;
      isInit = true;
    });
  }

  /**
   * Get title class name.
   * @param {Number} trend
   * @param {Number} stamp
   * @return {String}
   */
  function getClassNameByTrend(trend, stamp) {
    var isOutDate = false;
    var nowStamp = new Date() / 1000;

    if (nowStamp - 2 * $scope.filter.interval > stamp) {
      isOutDate = true;
    }

    if (Math.abs(trend) >= 1 && !isOutDate) {
      return 'anomalous';
    }
    return 'normal';
  }

  /**
   * Scrollbars
   */
  function initScrollbars() {
    $('.chart-box-top').scroll(function() {
      $('.chart-box').scrollLeft($('.chart-box-top').scrollLeft());
    });
    $('.chart-box').scroll(function() {
      $('.chart-box-top').scrollLeft($('.chart-box').scrollLeft());
    });
  }

  /**
   * Feed metric.
   * @param {String} name
   * @param {Function} cb // function(data)
   * @return {Metric}
   */
  function feed(name, data, cb) {
    return chart.metric(function(start, stop, step, callback) {
      var values = [], i = 0;
      // cast to timestamp from date
      start = parseInt((+start - $scope.filter.datetime) / 1000);
      stop = parseInt((+stop - $scope.filter.datetime) / 1000);
      step = parseInt(+step / 1000);
      // parameters to pull data
      var params = {name: name, start: start, stop: stop};
      // request data and call `callback` with values
      // data schema: {name: {String}, times: {Array}, vals: {Array}}
      Metric.getMetricValues(params, function(data) {
        // the timestamps from statsd DONT have exactly steps `10`
        var len = data.length;
        while (start < stop && i < len) {
          while (start < data[i].stamp) {
            start += step;
            if ($scope.filter.type === 'v') {
              values.push(start > data[i].stamp ? data[i].value : 0);
            } else {
              values.push(start > data[i].stamp ? data[i].score : 0);
            }
          }

          if ($scope.filter.type === 'v') {
            values.push(data[i++].value);
          } else {
            values.push(data[i++].score);
          }
          start += step;
        }
        callback(null, values);

      });
      cb(data);
    }, name);
  }

  /**
   * Get trend text.
   * @param {Number} trend
   * @return {String}
   */
  function getTextByTrend(trend) {
    if (trend > 0) {
      return '↑ ';
    }

    if (trend < 0) {
      return '↓ ';
    }

    return '- ';
  }

  function setIntervalAndRunNow(fn, ms) {
    fn();
    return setInterval(fn, ms);
  }

  $scope.isGraphiteName = Util.isGraphiteName;
  $scope.translateGraphiteName = Util.translateGraphiteName;

  $scope.datetimeRangeInString = function() {
    if (!chart.context()) {
      return '';
    }
    var stop = +new Date() - $scope.filter.datetime * 1000;
    var start = stop - chart.size() * chart.step();
    return Util.format('%s ~ %s', Util.dateToString(start),
                       Util.dateToString(stop));
  };

  $scope.datetimeInList = function(seconds) {
    var i;
    for (i = 0; i < $scope.dateTimes.length; i++) {
      if ($scope.dateTimes[i].seconds === seconds) {
        return true;
      }
    }
    return false;
  };

  $scope.secondsToTimespanString = Util.secondsToTimespanString;

  loadData();
  initScrollbars();
};
