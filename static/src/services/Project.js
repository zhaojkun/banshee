/**
 * Created by Panda on 16/1/14.
 */
/*@ngInject*/
module.exports = function($resource) {
  return $resource('/api/project/:id', {id: '@id', userId: '@userId'}, {
    edit: {method: 'PATCH', url: '/api/project/:id'},
    getAllProjects: {method: 'GET', url: '/api/projects', isArray: true},
    getRulesByProjectId:
        {method: 'GET', url: '/api/project/:id/rules', isArray: true},
    getUsersByProjectId:
        {method: 'GET', url: '/api/project/:id/users', isArray: true},
    addUserToProject: {method: 'POST', url: '/api/project/:id/user'},
    deleteUserFromProject:
        {method: 'DELETE', url: '/api/project/:id/user/:userId'},
    getWebHooksByProjectId:
        {method: 'GET', url: '/api/project/:id/webhooks', isArray: true},
    addWebHookToProject: {method: 'POST', url: '/api/project/:id/webhook'},
    deleteWebHookFromProject:
        {method: 'DELETE', url: '/api/project/:id/webhook/:webhookId'},
    getEventsByProjectId:
        {method: 'GET', url: '/api/project/:id/events', isArray: true},
  });
};
