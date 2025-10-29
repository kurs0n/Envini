export SONAR_TOKEN=sqp_b1f07cdec80c317b56735a927d69b48770ac498d
docker run --rm \
  --platform linux/amd64 \
  --network envini_default \
  -e SONAR_HOST_URL="http://sonarqube:9000" \
  -e SONAR_TOKEN="$SONAR_TOKEN" \
  -v "$PWD:/usr/src" \
  sonarsource/sonar-scanner-cli