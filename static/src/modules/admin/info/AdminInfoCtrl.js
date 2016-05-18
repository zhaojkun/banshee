/*@ngInject*/
module.exports = function($scope, Info) {
  $scope.info = {};

  $scope.loadData = function() {
    Info.get().$promise.then(function(res) {
      if (Object.keys(res).length === 0) {
        $scope.info = null;
      } else {
        // Tofixed with cost
        res.detectionCost = res.detectionCost.toFixed(4);
        res.filterCost = res.filterCost.toFixed(4);
        res.queryCost = res.queryCost.toFixed(4);
        res.metricCacheInitCost = res.metricCacheInitCost.toFixed(4);
        $scope.info = res;
      }
    });
  };

  $scope.loadData();
};
