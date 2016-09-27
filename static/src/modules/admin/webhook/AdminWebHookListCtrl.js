/*@ngInject*/
module.exports = function ($scope, $modal, $mdDialog, $state, $timeout, WebHook) {
  $scope.autoComplete = {
    searchText: ''
  };

  $scope.loadData = function () {
    WebHook.getAllWebHooks().$promise
      .then(function (res) {
        $scope.webhooks = res;
      });
  };

  $scope.searchWebHook = function (webhook) {
    console.log(webhook)
    $timeout(function () {
      $state.go('banshee.admin.webhook.detail', {
        id: webhook.id
      });
    }, 200);
  };

  $scope.openModal = function (event) {
    $mdDialog.show({
        controller: 'WebHookAddModalCtrl',
        templateUrl: 'modules/admin/webhook/webHookAddModal.html',
        parent: angular.element(document.body),
        targetEvent: event,
        clickOutsideToClose: true,
        fullscreen: true
      })
      .then(function (res) {
        $scope.webhooks.push(res);
      });
  };

  $scope.loadData();

};
