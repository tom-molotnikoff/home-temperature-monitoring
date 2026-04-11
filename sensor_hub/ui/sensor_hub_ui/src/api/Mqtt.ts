import { get, post, put, del, type ApiMessage } from './Client';
import type { MQTTBroker, MQTTSubscription } from '../types/types';

export type CreateBrokerPayload = {
  name: string;
  type: string;
  host: string;
  port: number;
  username?: string;
  password?: string;
  client_id?: string;
  enabled: boolean;
};

export type CreateSubscriptionPayload = {
  broker_id: number;
  topic_pattern: string;
  driver_type: string;
  enabled: boolean;
};

export const MqttBrokersApi = {
  list:   ()                                        => get<MQTTBroker[]>('/mqtt/brokers'),
  getById:(id: number)                              => get<MQTTBroker>(`/mqtt/brokers/${id}`),
  create: (broker: CreateBrokerPayload)             => post<{ id: number }>('/mqtt/brokers', broker),
  update: (id: number, broker: CreateBrokerPayload) => put<ApiMessage>(`/mqtt/brokers/${id}`, broker),
  delete: (id: number)                              => del<ApiMessage>(`/mqtt/brokers/${id}`),
};

export const MqttSubscriptionsApi = {
  list:   ()                                                 => get<MQTTSubscription[]>('/mqtt/subscriptions'),
  getById:(id: number)                                       => get<MQTTSubscription>(`/mqtt/subscriptions/${id}`),
  create: (sub: CreateSubscriptionPayload)                   => post<{ id: number }>('/mqtt/subscriptions', sub),
  update: (id: number, sub: CreateSubscriptionPayload)       => put<ApiMessage>(`/mqtt/subscriptions/${id}`, sub),
  delete: (id: number)                                       => del<ApiMessage>(`/mqtt/subscriptions/${id}`),
};
