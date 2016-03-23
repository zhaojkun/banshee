/*@ngInject*/
module.exports = function ($resource) {
  return $resource('/api/config', {}, {
    getInterval: {
      method: 'GET',
      url: '/api/interval'
    },
    getLanguage: {
      method: 'GET',
      url: '/api/language'
    },
    getPrivateDocUrl: {
      method: 'GET',
      url: '/api/privateDocUrl'
    },
  });
};
