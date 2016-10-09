/*@ngInject*/
module.exports = function ($scope, $modal, $mdDialog, $state, $timeout, Project) {
  $scope.autoComplete = {
    searchText: ''
  };

  $scope.loadData = function () {
    Project.getAllProjects().$promise
      .then(function (res) {
        $scope.projects = res;
      });
  };

  $scope.searchProject = function (item) {
    $timeout(function() {
      $state.go('banshee.admin.team.project.detail', {id: item.id});
    }, 200);
  };

  $scope.openModal = function (event) {
    $mdDialog.show({
        controller: 'ProjectModalCtrl',
        templateUrl: 'modules/admin/project/projectModal.html',
        parent: angular.element(document.body),
        targetEvent: event,
        clickOutsideToClose: true,
        fullscreen: true,
        locals: {
          params: {
            opt: 'addProject'
          }
        }
      })
      .then(function (project) {
        $scope.projects.push(project);
      });
  };

  $scope.loadData();

};
