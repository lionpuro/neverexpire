alter table notifications
drop constraint uq_notifications_user_id_host_id_due;

alter table notifications
add constraint uq_notifications_host_id_due unique (host_id, due);
