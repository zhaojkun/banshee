/*@ngInject*/
module.exports = function($scope,$translate, toastr, $mdDialog, Team) {

  $scope.team = this.team || {};

  if(this.team){
    $scope.isEdit = true;
  }
  
  $scope.cancel = function() {
    $mdDialog.cancel();
  };
  
  $scope.submit = function() {
    if ($scope.isEdit) {
      $scope.edit();
    } else {
      $scope.create();      
    }
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

  $scope.edit = function() {
    Team.edit($scope.team).$promise.then(function(res) {
      $mdDialog.hide(res);
      toastr.success($translate.instant('SAVE_SUCCESS'));
    }).catch(function(err) {
      toastr.error(err.msg);
    });
  };
  
};
