/*@ngInject*/
module.exports = function($scope, $state, $stateParams, $translate, toastr,
                          $mdDialog, WebHook) {
  var webHookId = $stateParams.id;

  $scope.loadData = function() {
    // get webhooks
    WebHook.get({id: webHookId}).$promise.then(function(res) { $scope.webhook = res; });

  };

  $scope.edit = function() {
    WebHook.edit($scope.webhook).$promise.then(function(res) {
      $scope.webhook = res;
      toastr.success($translate.instant('SAVE_SUCCESS'));
    }).catch (function(err) { toastr.error(err.msg); });
  };

  $scope.loadWebHookProjectsDone = false;
  $scope.loadWebHookProjects = function() {
    if ($scope.loadWebHookProjectsDone) {
      return;
    }
    setTimeout(function() {
      // get projects by webhook id
      WebHook.getProjectsByWebHookId({id: webHookId}).$promise.then(function(res) {
        $scope.projects = res;
        $scope.loadWebHookProjectsDone = true;
      });

    }, 500);
  };

  $scope.deleteWebHook = function(event) {
    var confirm = $mdDialog.confirm()
                      .title($translate.instant('ADMIN_WEBHOOK_DELETE_TITLE'))
                      .textContent($translate.instant('ADMIN_WEBHOOK_DELETE_WARN'))
                      .ariaLabel($translate.instant('ADMIN_WEBHOOK_DELETE_TEXT'))
                      .targetEvent(event)
                      .ok($translate.instant('YES'))
                      .cancel($translate.instant('NO'));
    $mdDialog.show(confirm).then(function() {
      WebHook.delete ({id: $scope.webhook.id}).$promise.then(function() {
        toastr.success($translate.instant('DELETE_SUCCESS'));
        $state.go('banshee.admin.webhook');
      }).catch (function(err) { toastr.error(err.msg); });
    });

  };

  $scope.loadData();

};
