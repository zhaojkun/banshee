/*@ngInject*/
module.exports = function($scope, $modal, $mdDialog, $state, $timeout, Team) {
  $scope.autoComplete = {
    searchText: ''
  };

  $scope.loadData = function() {
    Team.getAllTeams().$promise
      .then(function(res) {
        $scope.teams = res;
      });
  };
  
  $scope.searchTeam = function(item) {
    $timeout(function() {
      $state.go('banshee.admin.team.detail', {
        id: item.id
      });
    }, 200);
  };
  
  $scope.openModal = function(event) {
    $mdDialog.show({
      controller: 'TeamModalCtrl',
      templateUrl: 'modules/admin/team/teamModal.html',
      parent: angular.element(document.body),
      targetEvent: event,
      clickOutsideToClose: true,
      fullscreen: true,
    }).then(function(team) {
      $scope.teams.push(team);
    });
  };
  $scope.loadData();
};
