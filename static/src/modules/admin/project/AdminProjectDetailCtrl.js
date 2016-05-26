/*@ngInject*/
module.exports =
    function($scope, $location, $mdDialog, $state, $stateParams, $translate,
             toastr, Project, Rule, User, Config, Util) {
  var projectId = $scope.projectId = $stateParams.id;
  var allUsers = [];

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

    // get users of project
    Project.getUsersByProjectId({id: projectId})
        .$promise.then(function(res) { $scope.users = res; });

    // get all users
    User.getAllUsers().$promise.then(function(res) { allUsers = res; });

    // get config
    Config.get().$promise.then(function(res) { $scope.config = res; });

    // get events
    Project.getEventsByProjectId({id: projectId})
        .$promise.then(function(res) { $scope.events = res; });
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
        $state.go('banshee.admin.project');
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
    var ctrl, template, users;

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
                {opt: opt, obj: angular.copy(project) || '', users: users}
          }
        })
        .then(function(res) {
      if (opt === 'addRule') {
        $scope.rules.push(res);
      }
      if (opt === 'addUserToProject') {
        $scope.users.push(res);
      }
    });
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
                               0)))
  };

  $scope.loadData();

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

  $scope.buildRepr = Util.buildRepr;
  $scope.ruleCheck = Util.ruleCheck;
  $scope.dateToString = Util.dateToString;
  $scope.getEventRuleComment = function(event) {
    if (event.translatedComment.length > 0) return event.translatedComment;
    if (event.comment.length > 0) return event.comment;
    return event.pattern;
  };
  $scope.goToRuleID = function(ruleId) {
    $state.go('banshee.admin.project.detail',
              {id: $scope.projectId, rule: ruleId}, {reload: true});
  };
  $scope.goToMain = function(metricName, stamp) {
    var past = +new Date() / 1000 - stamp - 15 * 60;  // -15min
    $state.go('banshee.main',
              {pattern: metricName, past: Util.secondsToTimespanString(past)});
  };
};
