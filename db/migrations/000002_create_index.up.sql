create index idx_domains_user_id on domains(user_id);
create index idx_domains_expires_at on domains(expires_at);

create index idx_notifications_user_id on notifications(user_id);
create index idx_notifications_due on notifications(due);
create index idx_notifications_delivered_at on notifications(delivered_at);
