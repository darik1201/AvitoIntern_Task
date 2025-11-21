import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '30s', target: 10 },
    { duration: '1m', target: 50 },
    { duration: '30s', target: 100 },
    { duration: '1m', target: 100 },
    { duration: '30s', target: 50 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'],
    http_req_failed: ['rate<0.01'],
    errors: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export function setup() {
  const teams = [];
  
  for (let i = 1; i <= 5; i++) {
    const teamData = JSON.stringify({
      team_name: `team${i}`,
      members: [
        { user_id: `u${i}1`, username: `User${i}1`, is_active: true },
        { user_id: `u${i}2`, username: `User${i}2`, is_active: true },
        { user_id: `u${i}3`, username: `User${i}3`, is_active: true },
      ],
    });

    const res = http.post(`${BASE_URL}/team/add`, teamData, {
      headers: { 'Content-Type': 'application/json' },
    });

    if (res.status === 201) {
      teams.push(`team${i}`);
    }
  }

  return { teams };
}

export default function (data) {
  if (!data || !data.teams || data.teams.length === 0) {
    sleep(1);
    return;
  }
  
  const team = data.teams[Math.floor(Math.random() * data.teams.length)];
  if (!team) {
    sleep(1);
    return;
  }
  
  const userId = `u${team.replace('team', '')}1`;

  const prData = JSON.stringify({
    pull_request_id: `pr-${Date.now()}-${Math.random()}`,
    pull_request_name: 'Load test PR',
    author_id: userId,
  });

  const res = http.post(`${BASE_URL}/pullRequest/create`, prData, {
    headers: { 'Content-Type': 'application/json' },
  });

  const success = check(res, {
    'status is 201': (r) => r.status === 201,
    'response time < 300ms': (r) => r.timings.duration < 300,
  });

  errorRate.add(!success);

  sleep(0.1);
}

export function teardown(data) {
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, {
    'health check ok': (r) => r.status === 200,
  });

  const statsRes = http.get(`${BASE_URL}/stats`);
  check(statsRes, {
    'stats available': (r) => r.status === 200,
  });
}
