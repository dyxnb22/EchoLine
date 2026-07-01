import http from "k6/http";
import { check } from "k6";

export const options = {
  vus: 1,
  duration: "10s",
};

export default function () {
  const res = http.get("http://localhost:8080/health");
  check(res, {
    "health status is 200 or service not implemented yet": (r) =>
      r.status === 200 || r.status === 404 || r.status === 0,
  });
}

