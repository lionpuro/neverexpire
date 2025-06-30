alter table domains
drop column status;
alter table domains
add status int not null default 0;
