var app =
    angular.module('banshee.admin', [])
        /*@ngInject*/
        .config(
          function($stateProvider) {
               // State
               $stateProvider.state('banshee.admin', {
                                      url: '/admin',
                                           template: '<ui-view></ui-view>',
                                      abstract: true
                                    })

                 // Project router
                 .state('banshee.admin.team',{
                   url: '/team',
                   templateUrl: "modules/admin/team/list.html",
                   controller: "AdminTeamListCtrl"
                 })
                 .state('banshee.admin.team.detail',{
                   url: '/:id',
                   views: {
                     '@banshee':{
                       templateUrl: 'modules/admin/team/AdminTeamDetail.html',
                       controller: 'AdminTeamDetailCtrl'
                     }
                   }
                 })
                 .state('banshee.admin.team.project', {
                   url: '/project',
                   templateUrl: 'modules/admin/project/list.html',
                   controller: 'AdminProjectListCtrl'
                 })
                 .state('banshee.admin.team.project.detail', {
                   url: '/:id?rule',
                   views: {
                     '@banshee': {
                       templateUrl:
                         'modules/admin/project/AdminProjectDetail.html',
                       controller: 'AdminProjectDetailCtrl'
                       }
                   }
                 })
                 
                 // User router
                 .state('banshee.admin.user', {
                   url: '/user',
                   templateUrl: 'modules/admin/user/AdminUserList.html',
                   controller: 'AdminUserListCtrl'
                 })
                 .state('banshee.admin.user.detail', {
                   url: '/:id',
                   views: {
                     '@banshee': {
                       templateUrl: 'modules/admin/user/AdminUserDetail.html',
                       controller: 'AdminUserDetailCtrl'
                     }
                   }
                 })
                 
                 // Config router
                 .state('banshee.admin.config', {
                   url: '/config',
                   templateUrl: 'modules/admin/config/config.html',
                   controller: 'AdminConfigCtrl'
                 })
                 
                 // Info router
                 .state('banshee.admin.info', {
                   url: '/info',
                   templateUrl: 'modules/admin/info/info.html',
                   controller: 'AdminInfoCtrl'
                 });
             })

  // Controller
      .controller('AdminTeamListCtrl',require('./team/AdminTeamListCtrl'))
      .controller('AdminTeamDetailCtrl',require('./team/AdminTeamDetailCtrl'))
      .controller('TeamModalCtrl',require('./team/TeamModalCtrl'))
      .controller('ProjectModalCtrl', require('./team/ProjectModalCtrl'))
  
      .controller('AdminProjectListCtrl',
        require('./project/AdminProjectListCtrl'))
      .controller('AdminProjectDetailCtrl',
        require('./project/AdminProjectDetailCtrl'))
      .controller('UserModalCtrl', require('./project/UserModalCtrl'))
      .controller('RuleModalCtrl', require('./project/RuleModalCtrl'))
  
      .controller('AdminUserListCtrl', require('./user/AdminUserListCtrl'))
      .controller('AdminUserDetailCtrl',
        require('./user/AdminUserDetailCtrl'))
      .controller('UserAddModalCtrl', require('./user/UserAddModalCtrl'))
  
      .controller('AdminConfigCtrl', require('./config/AdminConfigCtrl'))
      .controller('AdminInfoCtrl', require('./info/AdminInfoCtrl'));

module.exports = app.name;
