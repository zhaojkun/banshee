/**
 * Created by Panda on 16/1/14.
 */
/*@ngInject*/
module.exports = function($resource) {
    return $resource('/api/team/:id', {
        id: '@id'
    }, {
        edit: {
            method: 'PATCH',
            url: '/api/team/:id'
        },
        getAllTeams: {
            method: 'GET',
            url: '/api/teams',
            isArray: true
        },
        getProjectsByTeamId: {
            method: 'GET',
            url: '/api/team/:id/projects',
            isArray: true
        },
    });
};