FROM ehazlett/mkdocs:latest as BUILD
COPY . /app
WORKDIR /app
RUN mkdocs build -d /app/_site --clean

FROM nginx:alpine
COPY --from=BUILD /app/_site /usr/share/nginx/html