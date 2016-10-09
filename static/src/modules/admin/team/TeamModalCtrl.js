/*@ngInject*/
module.exports = function($scope, toastr, $mdDialog, Team) {

  $scope.team = {};
  
  $scope.cancel = function() {
    $mdDialog.cancel();
  };
  
  $scope.submit = function() {
    $scope.create();
  };
  
  $scope.create = function() {
    Team.save($scope.team).$promise
      .then(function(res) {
        $mdDialog.hide(res);
      })
      .catch(function(err) {
        toastr.error(err.msg);
      });
  };
};
