/*@ngInject*/
module.exports =
  function($scope, $location, $mdDialog, $state, $stateParams, $translate,
    toastr, Project, Rule, User, Config, Util,Team,WebHook) {
    var projectId = $scope.projectId = $stateParams.id;
    var allUsers = [];
    var allWebHooks=[];

    $scope.loadData = function() {

      // get project
      Project.get({id: $stateParams.id})
        .$promise.then(function(res) { $scope.project = res; });

      // get rules of project
      Project.getRulesByProjectId({id: projectId}).$promise.then(function(res) {
        $scope.rules = res;
        // Open modal to edit rule if 'rule' is in url query strings
        var queryStringRuleID = $location.search().rule;
        if (queryStringRuleID) {
          for (var i = 0; i < $scope.rules.length; i++) {
            if ($scope.rules[i].id === +queryStringRuleID) {
              $scope.editRule(null, $scope.rules[i]);
              break;
            }
          }
        }
      });

      Team.get({id: $stateParams.teamID})
        .$promise.then(function(res) {$scope.team = res;});
      
      Team.getAllTeams().$promise
        .then(function(res) {
          $scope.teams = res;
        });

      // get config
      Config.get().$promise.then(function(res) { $scope.config = res; });

    };

    $scope.loadUsersDone = false;
    $scope.loadUsers = function() {
      if ($scope.loadUsersDone) {
        return;
      }
      setTimeout(function() {
        // get users of project
        Project.getUsersByProjectId({id: projectId})
          .$promise.then(function(res) { $scope.users = res; });

        // get all users
        User.getAllUsers().$promise.then(function(res) { allUsers = res; });
        $scope.loadUsersDone = true;
      }, 500);
    };

    $scope.loadWebHooksDone = false;
    $scope.loadWebHooks = function() {
      if ($scope.loadWebHooksDone) {
        return;
      }
      setTimeout(function() {
        // get users of project
        Project.getWebHooksByProjectId({id: projectId})
          .$promise.then(function(res) { $scope.webhooks= res; });

        // get all users
        WebHook.getAllWebHooks().$promise.then(function(res) { allWebHooks = res; });
        $scope.loadWebHooksDone = true;
      }, 500);
    };


    
    $scope.edit = function() {
      Project.edit($scope.project).$promise.then(function() {
        toastr.success($translate.instant('SAVE_SUCCESS'));
      }).catch (function(err) { toastr.error(err.msg); });
    };

    $scope.deleteRule = function(event, ruleId, index) {
      var confirm = $mdDialog.confirm()
        .title($translate.instant('ADMIN_RULE_DELETE_TEXT'))
        .textContent($translate.instant('ADMIN_RULE_DELETE_WARN'))
        .ariaLabel($translate.instant('ADMIN_RULE_DELETE_TEXT'))
        .targetEvent(event)
        .ok($translate.instant('YES'))
        .cancel($translate.instant('NO'));
      $mdDialog.show(confirm).then(function() {
        Rule.delete ({id: ruleId}).$promise.then(function() {
          $scope.rules.splice(index, 1);
          toastr.success($translate.instant('DELETE_SUCCESS'));
        }).catch (function(err) { toastr.error(err.msg); });
      });
    };

    $scope.deleteUser = function(event, userId, index) {
      var confirm = $mdDialog.confirm()
        .title($translate.instant('ADMIN_USER_REMOVE_TEXT'))
        .textContent($translate.instant('ADMIN_USER_REMOVE_WARN'))
        .ariaLabel($translate.instant('ADMIN_USER_REMOVE_TEXT'))
        .targetEvent(event)
        .ok($translate.instant('YES'))
        .cancel($translate.instant('NO'));
      $mdDialog.show(confirm).then(function() {
        Project.deleteUserFromProject({id: projectId, userId: userId})
          .$promise.then(function() {
            $scope.users.splice(index, 1);
            toastr.success($translate.instant('DELETE_SUCCESS'));
          }).catch (function(err) { toastr.error(err.msg); });
      });
    };

    $scope.deleteWebHook = function(event, webhookId, index) {
      var confirm = $mdDialog.confirm()
        .title($translate.instant('ADMIN_WEBHOOK_REMOVE_TEXT'))
        .textContent($translate.instant('ADMIN_WEBHOOK_REMOVE_WARN'))
        .ariaLabel($translate.instant('ADMIN_WEBHOOK_REMOVE_TEXT'))
        .targetEvent(event)
        .ok($translate.instant('YES'))
        .cancel($translate.instant('NO'));
      $mdDialog.show(confirm).then(function() {
        Project.deleteWebHookFromProject({id: projectId, webhookId: webhookId})
          .$promise.then(function() {
            $scope.webhooks.splice(index, 1);
            toastr.success($translate.instant('DELETE_SUCCESS'));
          }).catch (function(err) { toastr.error(err.msg); });
      });
    };

    $scope.deleteProject = function(event) {
      var confirm =
        $mdDialog.confirm()
          .title($translate.instant('ADMIN_PROJECT_DELETE_TEXT'))
          .textContent($translate.instant('ADMIN_PROJECT_DELETE_WARN'))
          .ariaLabel($translate.instant('ADMIN_PROJECT_DELETE_TEXT'))
          .targetEvent(event)
          .ok($translate.instant('YES'))
          .cancel($translate.instant('NO'));

      $mdDialog.show(confirm).then(function() {
        Project.delete ({id: $scope.project.id}).$promise.then(function() {
          toastr.success($translate.instant('DELETE_SUCCESS'));
          $state.go('banshee.admin.team.project');
        }).catch (function(err) { toastr.error(err.msg); });
      });

    };

    $scope.editRule = function(event, rule) {
      $mdDialog.show({
        controller: 'RuleModalCtrl',
        templateUrl: 'modules/admin/project/ruleModal.html',
        parent: angular.element(document.body),
        targetEvent: event,
        clickOutsideToClose: true,
        fullscreen: true,
        bindToController: true,
        locals: {
          rule: rule,
        }
      });
    };

    $scope.openModal = function(event, opt, project) {
      var ctrl, template, users,webhooks;

      if (opt === 'addRule') {
        ctrl = 'RuleModalCtrl';
        template = 'modules/admin/project/ruleModal.html';
      }

      if (opt === 'editProject') {
        ctrl = 'ProjectModalCtrl';
        template = 'modules/admin/project/projectModal.html';
      }

      if (opt === 'addUserToProject') {
        ctrl = 'UserModalCtrl';
        template = 'modules/admin/project/userModal.html';
        users = filterUsers();
      }

      if (opt === 'addWebHookToProject'){
        ctrl = 'WebHookModalCtrl';
        template = 'modules/admin/project/webHookModal.html';
        webhooks = filterWebHooks();
      }
      
      $mdDialog
        .show({
          controller: ctrl,
          templateUrl: template,
          parent: angular.element(document.body),
          targetEvent: event,
          clickOutsideToClose: true,
          fullscreen: true,
          locals: {
            params:
                {opt: opt, obj: angular.copy(project) || '', users: users,webhooks:webhooks}
          }
        })
        .then(function(res) {
          if (opt === 'addRule') {
            $scope.rules.push(res);
          }
          if (opt === 'addUserToProject') {
            $scope.users.push(res);
          }
          if (opt === 'addWebHookToProject') {
            $scope.webhooks.push(res);
          }

        });
    };
    
    $scope.exportRules = function(){
      Project.getRulesByProjectId({id: projectId}).$promise.then(function(res) {
        Util.saveContentToFile(JSON.stringify(res),'rules.json');
      });
    };
    
    $scope.importRules = function(file){
      if (file){
        Project.importRules({id: projectId,file:file}).$promise.then(function(res){
          $scope.loadData();
          var table = '<table><thead><tr><th><span>rule</span></th>'+
            '<th><span>error</span></th></tr></thead><tbody>';
          var found = false;
          for(var i=0;i<res.length;i++){
            if(res[i].Status!==null){
              found = true;
              table += '<tr><td>'+res[i].Rule+'</td>'+
                '<td>'+res[i].Status.msg+'</td></tr>';
            }
          }
          if(found){
            table += '</tbody></table>';
            var confirm = $mdDialog.alert()
              .title('errors')
              .htmlContent(table)
              .ok($translate.instant('OK'));
            $mdDialog.show(confirm);
          }
        });
      }
  };


  $scope.translateRuleRepr = function(rule) {
    return Util.translateRuleRepr(rule, $scope.config, $translate);
  };

  $scope.translateRuleNumMetricsWarn = function() {
    return $translate.instant('ADMIN_RULE_NUM_METRICS_WARN', {
      interval: $scope.config.interval,
      threshold: $scope.config.detector.intervalHitLimit
    });
  };

  // Returns true if the rule is disabled forever or disabled at this time
  // temply.
  $scope.isRuleDisabledWorksNow = function(rule) {
    return rule.disabled && (rule.disabledFor <= 0 ||
                             (rule.disabledFor > 0 &&
                              (+new Date(rule.disabledAt) / 1000 +
                                   rule.disabledFor * 60 - new Date() / 1000 >
                               0)));
  };

  /**
   * filter user:
   *  1.user.universal = true;
   *  2.user is not the existing user list;
   * @param
   */
  function filterUsers() {
    var usersIds = getUsersId();
    return allUsers.map(function(el) {
      if (!el.universal && usersIds.indexOf(el.id) < 0) {
        return el;
      }
    });
  }

  function getUsersId() {
    return $scope.users.map(function(el) { return el.id; });
  }

  function filterWebHooks() {
    var webhooksIds = getWebHooksId();
    return allWebHooks.map(function(el) {
      if (webhooksIds.indexOf(el.id) < 0) {
        return el;
      }
    });
  }
  function getWebHooksId() {
    return $scope.webhooks.map(function(el) { return el.id; });
  }


      
  $scope.buildRepr = Util.buildRepr;
  $scope.ruleCheck = Util.ruleCheck;
  $scope.dateToString = Util.dateToString;
  $scope.getEventRuleComment = function(event) {
    if (event.translatedComment.length > 0) {
      return event.translatedComment;
    }
    if (event.comment.length > 0) {
      return event.comment;
    }
    return event.pattern;
  };
  $scope.goToRuleID = function(ruleId) {
    $state.go('banshee.admin.team.project.detail',
              {id: $scope.projectId, rule: ruleId}, {reload: true});
  };
  $scope.goToMetric = function(metricName, stamp) {
    var past = +new Date() / 1000 - stamp - 15 * 60;  // -15min
    $state.go('banshee.metric',
              {pattern: metricName, past: Util.secondsToTimespanString(past)});
  };

  $scope.eventPasts = [
    {label: 'EVENT_PAST_1DAY', seconds: 3600 * 24 * 1},
    {label: 'EVENT_PAST_2DAYS', seconds: 3600 * 24 * 2},
    {label: 'EVENT_PAST_3DAYS', seconds: 3600 * 24 * 3},
    {label: 'EVENT_PAST_4DAYS', seconds: 3600 * 24 * 4},
    {label: 'EVENT_PAST_5DAYS', seconds: 3600 * 24 * 5},
    {label: 'EVENT_PAST_6DAYS', seconds: 3600 * 24 * 6},
    {label: 'EVENT_PAST_7DAYS', seconds: 3600 * 24 * 7},
  ];

  $scope.eventPast = $scope.eventPasts[0].seconds;

  $scope.eventLevels = [
    {label: 'EVENT_LEVEL_LOW', value: 0},
    {label: 'EVENT_LEVEL_MIDDLE', value: 1},
    {label: 'EVENT_LEVEL_HIGH', value: 2},
  ];

  $scope.eventLevel = $scope.eventLevels[0].value;

  $scope.loadEvents = function() {
    Project
      .getEventsByProjectId(
             {id: projectId, past: $scope.eventPast, level: $scope.eventLevel})
      .$promise.then(function(res) { $scope.events = res; });
  };

  $scope.watchEventLoadParams = function() {
    if ($scope.watchEventLoadParamsDone) {
      return;
    }
    $scope.$watchGroup(['eventPast', 'eventLevel'], function() {
      $scope.events = null;
      setTimeout(function() { $scope.loadEvents(); }, 500);
    });
    $scope.watchEventLoadParamsDone = true;
  };

  $scope.watchEventLoadParamsDone = false;

  $scope.getGraphiteUrl = function(name) {
    var fmt = $scope.config.webapp.graphiteUrl;
    return Util.format(fmt, Util.getGraphiteName(name));
  };
      
  $scope.loadData();

  $scope.foldNumber = Util.foldNumber;
};
