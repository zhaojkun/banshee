/*@ngInject*/
module.exports = function ($scope, $mdDialog, $stateParams, $translate) {

  $scope.rule_file = null;
  
  $scope.cancel = function() {
    $mdDialog.cancel();
  };
  
  $scope.change = function() {
    console.log($scope.rule_file);
  };
  
  $scope.submit = function() {
    Project.addUserToProject({
      id: $stateParams.id,
      name: $scope.autoComplete.selectedItem.name
    }).$promise
    .then(function() {
      $mdDialog.hide($scope.autoComplete.selectedItem);
    })
    .catch(function(err) {
      toastr.error(err.msg);
    });
  };
};
