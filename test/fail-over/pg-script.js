import http from 'k6/http';
import { sleep } from 'k6';


export function pg_pghealth () {
    const url = 'https://ac7cd861bf3544dbabd81392c4a2ead8-2d20923fc4ccb0b6.elb.us-east-2.amazonaws.com/pg/pghealth'
    const params = {
        headers: {
            'Host' : 'us-east-2.pg-client.konghq.tech',
        },
    }
    http.get(url,params);
    sleep(1);
}

export function pg_replstatusro () {
    const url = 'https://ac7cd861bf3544dbabd81392c4a2ead8-2d20923fc4ccb0b6.elb.us-east-2.amazonaws.com/pg/replstatusro'
    const params = {
        headers: {
            'Host' : 'us-east-2.pg-client.konghq.tech',
        },
    }
    http.get(url,params);
    sleep(1);
}

export function pg_getFoo() {
    const url = 'https://ac7cd861bf3544dbabd81392c4a2ead8-2d20923fc4ccb0b6.elb.us-east-2.amazonaws.com/pg/foo'
    const params = {
        headers: {
            'Host' : 'us-east-2.pg-client.konghq.tech',
        },
    }
    http.get(url,params);
    sleep(1);
}

export function pg_postFoo() {
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
        pg_pghealth : {
            executor: 'constant-vus',
            exec: 'pg_pghealth',
            vus: 20,
            duration: '30s',
        },
        pg_replstatusro : {
            executor: 'constant-vus',
            exec: 'pg_replstatusro',
            vus: 20,
            duration: '30s',
        },
        pg_getFoo : {
            executor: 'constant-vus',
            exec: 'pg_getFoo',
            vus: 20,
            duration: '30s',
        },
        pg_postFoo: {
            executor: 'constant-vus',
            exec: 'pg_postFoo',
            vus: 5,
            duration: '30s',
        },
    },
};
