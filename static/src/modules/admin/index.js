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
                   .state('banshee.admin.project', {
                     url: '/project',
                     templateUrl: 'modules/admin/project/list.html',
                     controller: 'AdminProjectListCtrl'
                   })
                   .state('banshee.admin.project.detail', {
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


	       // WebHook router
                   .state('banshee.admin.webhook', {
                     url: '/webhook',
                     templateUrl: 'modules/admin/webhook/AdminWebHookList.html',
                     controller: 'AdminWebHookListCtrl'
                   })
                   .state('banshee.admin.webhook.detail', {
                     url: '/:id',
                     views: {
                       '@banshee': {
                         templateUrl: 'modules/admin/webhook/AdminWebHookDetail.html',
                         controller: 'AdminWebHookDetailCtrl'
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
        .controller('AdminProjectListCtrl',
                    require('./project/AdminProjectListCtrl'))
        .controller('AdminProjectDetailCtrl',
                    require('./project/AdminProjectDetailCtrl'))
        .controller('ProjectModalCtrl', require('./project/ProjectModalCtrl'))
        .controller('UserModalCtrl', require('./project/UserModalCtrl'))
        .controller('WebHookModalCtrl', require('./project/WebHookModalCtrl'))
        .controller('RuleModalCtrl', require('./project/RuleModalCtrl'))

        .controller('AdminUserListCtrl', require('./user/AdminUserListCtrl'))

        .controller('AdminUserDetailCtrl',
                    require('./user/AdminUserDetailCtrl'))

        .controller('UserAddModalCtrl', require('./user/UserAddModalCtrl'))

        .controller('AdminWebHookListCtrl', require('./webhook/AdminWebHookListCtrl'))
        .controller('AdminWebHookDetailCtrl',
                    require('./webhook/AdminWebHookDetailCtrl'))

        .controller('WebHookAddModalCtrl', require('./webhook/WebHookAddModalCtrl'))


        .controller('AdminConfigCtrl', require('./config/AdminConfigCtrl'))
        .controller('AdminInfoCtrl', require('./info/AdminInfoCtrl'));

module.exports = app.name;
