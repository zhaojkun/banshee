/**
 * Created by Panda on 16/1/16.
 */
/*@ngInject*/
module.exports = function ($resource) {
  return $resource('/api/webhook/:id', {projectId: '@projectId', id: '@id'}, {
    getAllWebHooks: {
      method: 'GET',
      url: '/api/webhooks',
      isArray: true
    },
    getProjectsByWebHookId: {
      method: 'GET',
      url: '/api/webhook/:id/projects',
      isArray: true
    },
    edit: {
      method: 'PATCH',
      url: '/api/webhook/:id'
    }
  });
};
