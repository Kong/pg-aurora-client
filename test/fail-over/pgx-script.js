import http from 'k6/http';
import { sleep } from 'k6';

export function pgx_pghealth () {
    const url = 'https://k8s-kong-kongkong-222619ec7a-13992fea9f16a2fc.elb.us-east-2.amazonaws.com/pgx/pghealth'
    const params = {
        headers: {
            'Host' : 'us-east-2.pg-client.konghq.tech',
        },
    }
    http.get(url,params);
    sleep(0.05);
}

export function pgx_replstatusro () {
    const url = 'https://k8s-kong-kongkong-222619ec7a-13992fea9f16a2fc.elb.us-east-2.amazonaws.com/pgx/ro/replstatus'
    const params = {
        headers: {
            'Host' : 'us-east-2.pg-client.konghq.tech',
        },
    }
    http.get(url,params);
    sleep(0.05);
}

export function pgx_getFoo() {
    const url = 'https://k8s-kong-kongkong-222619ec7a-13992fea9f16a2fc.elb.us-east-2.amazonaws.com/pgx/foo'
    const params = {
        headers: {
            'Host' : 'us-east-2.pg-client.konghq.tech',
        },
    }
    http.get(url,params);
    sleep(0.05);
}

export function pgx_postFoo() {
    const url = 'https://k8s-kong-kongkong-222619ec7a-13992fea9f16a2fc.elb.us-east-2.amazonaws.com/pgx/foo'
    const params = {
        headers: {
            'Host' : 'us-east-2.pg-client.konghq.tech',
        },
    }
    http.post(url,null,params);
    sleep(0.05);
}

export const options = {
    insecureSkipTLSVerify: true,
    scenarios: {
        pgx_pghealth: {
            executor: 'constant-vus',
            exec: 'pgx_pghealth',
            vus: 5,
            duration: '600s',
        },
        pgx_replstatusro: {
            executor: 'constant-vus',
            exec: 'pgx_replstatusro',
            vus: 5,
            duration: '600s',
        },
        pgx_getFoo: {
            executor: 'constant-vus',
            exec: 'pgx_getFoo',
            vus: 5,
            duration: '600s',
        },
        pgx_postFoo: {
            executor: 'constant-vus',
            exec: 'pgx_postFoo',
            vus: 15,
            duration: '600s',
        },
    },

};
