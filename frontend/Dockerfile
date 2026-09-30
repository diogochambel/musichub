FROM node:20-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

RUN chmod +x entrypoint.sh

EXPOSE 4200

ENTRYPOINT ["./entrypoint.sh"]