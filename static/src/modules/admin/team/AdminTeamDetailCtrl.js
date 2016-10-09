/*@ngInject*/
module.exports =
function($scope, $location, $mdDialog, $state, $stateParams, $translate,
        toastr, Project, Rule, User, Config, Util, Team) {
          var teamId = $scope.teamId = $stateParams.id;

          $scope.loadData = function() {
            // get project
            Team.get({
              id: $stateParams.id
            })
              .$promise.then(function(res) {
                $scope.team = res;
              });
            
            // get projects of team
            Team.getProjectsByTeamId({
              id: teamId
            })
              .$promise.then(function(res) {
                $scope.projects = res;
            });
            
          };
          
          $scope.edit = function() {
            Team.edit($scope.team).$promise.then(function() {
              toastr.success($translate.instant('SAVE_SUCCESS'));
            }).catch(function(err) {
              toastr.error(err.msg);
            });
          };
          
          $scope.deleteTeam = function(event) {
            var confirm =
            $mdDialog.confirm()
              .title($translate.instant('ADMIN_PROJECT_DELETE_TEXT'))
              .textContent($translate.instant('ADMIN_PROJECT_DELETE_WARN'))
              .ariaLabel($translate.instant('ADMIN_PROJECT_DELETE_TEXT'))
              .targetEvent(event)
              .ok($translate.instant('YES'))
              .cancel($translate.instant('NO'));
            $mdDialog.show(confirm).then(function() {
              Team.delete({
                id: $scope.team.id
              }).$promise.then(function() {
                toastr.success($translate.instant('DELETE_SUCCESS'));
                $state.go('banshee.admin.team');
              }).catch(function(err) {
                toastr.error(err.msg);
              });
            });
          };

          $scope.openModal = function (event) {
            $mdDialog.show({
              controller: 'ProjectModalCtrl',
              templateUrl: 'modules/admin/team/projectModal.html',
              parent: angular.element(document.body),
              targetEvent: event,
              clickOutsideToClose: true,
              fullscreen: true,
            })
              .then(function (project) {
                $scope.projects.push(project);
              });
          };
          
          $scope.loadData();
          $scope.foldNumber = Util.foldNumber;
};
