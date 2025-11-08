export SONAR_TOKEN=sqp_9ae91129b96dfc534769cf378b4b1d6a60f17737
docker run --rm \
  --platform linux/amd64 \
  --network envini_default \
  -e SONAR_HOST_URL="http://sonarqube:9000" \
  -e SONAR_TOKEN="$SONAR_TOKEN" \
  -v "$PWD:/usr/src" \
  sonarsource/sonar-scanner-cli