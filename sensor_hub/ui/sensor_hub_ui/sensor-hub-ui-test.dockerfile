FROM node:20 AS build
WORKDIR /app
COPY package*.json ./

COPY . .
RUN npm install
ARG VITE_API_BASE=http://localhost:8080
ARG VITE_WEBSOCKET_BASE=ws://localhost:8080
ENV VITE_API_BASE=$VITE_API_BASE
ENV VITE_WEBSOCKET_BASE=$VITE_WEBSOCKET_BASE

CMD ["npm", "run", "dev", "--", "--host"]