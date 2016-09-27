/*@ngInject*/
module.exports = function ($scope, $mdDialog, $stateParams, $translate, toastr, Project, params) {
  $scope.titles = {
    addWebHookToProject: $translate.instant('ADMIN_WEBHOOK_ADD_TEXT')
  };
  $scope.opt = params.opt;

  $scope.webhooks = params.webhooks;

  $scope.autoComplete = {
    searchText: ''
  };


  $scope.cancel = function() {
    $mdDialog.cancel();
  };

  $scope.submit = function() {
    Project.addWebHookToProject({
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
