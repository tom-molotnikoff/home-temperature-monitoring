FROM node:20 AS build
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
ARG VITE_API_BASE=http://localhost:8080
ARG VITE_WEBSOCKET_BASE=ws://localhost:8080
ENV VITE_API_BASE=$VITE_API_BASE
ENV VITE_WEBSOCKET_BASE=$VITE_WEBSOCKET_BASE
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
# copy custom nginx config
COPY ./nginx/default.conf /etc/nginx/conf.d/default.conf

# ensure certificate paths exist for docker-compose mounts
RUN mkdir -p /etc/ssl/certs /etc/ssl/private
EXPOSE 80 443
CMD ["nginx", "-g", "daemon off;"]