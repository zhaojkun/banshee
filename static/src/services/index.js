/**
 * Created by Panda on 16/1/13.
 */

var app = angular.module('banshee.services', ['ngResource']);

app.config(function ($httpProvider) {
  $httpProvider.interceptors.push('httpInterceptor');
});

app
  .factory('httpInterceptor', require('./httpInterceptor'))
  .factory('Project', require('./Project'))
  .factory('User', require('./User'))
  .factory('WebHook', require('./WebHook'))
  .factory('Rule', require('./Rule'))
  .factory('Config', require('./Config'))
  .factory('Metric', require('./Metric'))
  .factory('Info', require('./Info'))
  .factory('Version', require('./Version'))
  .factory('Util', require('./Util'))
  ;

module.exports = app.name;
