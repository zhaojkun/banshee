var app = angular.module('banshee.metric', [])
              /*@ngInject*/
              .config(
                   function($stateProvider) {

                     // State
                     $stateProvider.state('banshee.metric', {
                       url: '/metric?pattern&project&past',
                       templateUrl: 'modules/metric/list.html',
                       controller: 'MetricListCtrl'
                     });
                   })

              // Controller
              .controller('MetricListCtrl', require('./MetricListCtrl'));

module.exports = app.name;
