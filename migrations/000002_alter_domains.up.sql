alter table domains
add constraint unique_domain_per_user unique (user_id, domain_name);
