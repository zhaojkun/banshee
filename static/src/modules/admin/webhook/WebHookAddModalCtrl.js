/*@ngInject*/
module.exports = function ($scope, $mdDialog, $stateParams, toastr, WebHook) {
  $scope.cancel = function() {
    $mdDialog.cancel();
  };

  $scope.submit = function() {
    WebHook.save($scope.webhook).$promise
      .then(function(res) {
        $mdDialog.hide(res);
      })
      .catch(function(err) {
        toastr.error(err.msg);
      });
  };
};
