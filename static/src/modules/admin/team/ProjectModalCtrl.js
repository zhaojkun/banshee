/*@ngInject*/
module.exports = function ($scope,$stateParams, toastr, $mdDialog, Project) {

  $scope.project = {};

  $scope.cancel = function() {
    $mdDialog.cancel();
  };

  $scope.submit = function() {
    $scope.create();
  };

  $scope.create = function() {
    var params = angular.copy($scope.project);
    params.teamId = $stateParams.id;
    Project.save(params).$promise
      .then(function(res) {
        $mdDialog.hide(res);
      })
      .catch(function(err) {
        toastr.error(err.msg);
      });
  };
};
