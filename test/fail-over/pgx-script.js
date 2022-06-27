import http from 'k6/http';
import { sleep } from 'k6';

export function pgx_pghealth () {
    const url = 'https://ac7cd861bf3544dbabd81392c4a2ead8-2d20923fc4ccb0b6.elb.us-east-2.amazonaws.com/pgx/pghealth'
    const params = {
        headers: {
            'Host' : 'us-east-2.pg-client.konghq.tech',
        },
    }
    http.get(url,params);
    sleep(1);
}

export function pgx_replstatusro () {
    const url = 'https://ac7cd861bf3544dbabd81392c4a2ead8-2d20923fc4ccb0b6.elb.us-east-2.amazonaws.com/pgx/replstatusro'
    const params = {
        headers: {
            'Host' : 'us-east-2.pg-client.konghq.tech',
        },
    }
    http.get(url,params);
    sleep(1);
}

export function pgx_getFoo() {
    const url = 'https://ac7cd861bf3544dbabd81392c4a2ead8-2d20923fc4ccb0b6.elb.us-east-2.amazonaws.com/pgx/foo'
    const params = {
        headers: {
            'Host' : 'us-east-2.pg-client.konghq.tech',
        },
    }
    http.get(url,params);
    sleep(1);
}

export function pgx_postFoo() {
    const url = 'https://ac7cd861bf3544dbabd81392c4a2ead8-2d20923fc4ccb0b6.elb.us-east-2.amazonaws.com/pgx/foo'
    const params = {
        headers: {
            'Host' : 'us-east-2.pg-client.konghq.tech',
        },
    }
    http.post(url,null,params);
    sleep(1);
}

export const options = {
    insecureSkipTLSVerify: true,
    scenarios: {
        pgx_pghealth: {
            executor: 'constant-vus',
            exec: 'pgx_pghealth',
            vus: 5,
            duration: '30s',
        },
        pgx_replstatusro: {
            executor: 'constant-vus',
            exec: 'pgx_replstatusro',
            vus: 5,
            duration: '30s',
        },
        pgx_getFoo: {
            executor: 'constant-vus',
            exec: 'pgx_getFoo',
            vus: 5,
            duration: '30s',
        },
        pgx_postFoo: {
            executor: 'constant-vus',
            exec: 'pgx_postFoo',
            vus: 5,
            duration: '30s',
        },
    },

};
