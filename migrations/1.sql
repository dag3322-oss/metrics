create table if not exists metric(id text primary key, "type" text not null, delta bigint, "value" float, hash text);

create or replace function metrics_get(i_params jsonb) returns jsonb language plpgsql as $$
declare v_id text;
begin
	v_id:= i_params ->> 'id';
	return coalesce(jsonb_agg(jsonb_strip_nulls(jsonb_build_object('id', id, 'type', "type", 'delta', delta, 'value', "value", 'hash', hash))), '[]') 
	from metric where id = coalesce(v_id, id);
end $$;

create or replace function metrics_set(i_list jsonb) returns void language plpgsql as $$
begin
	with a as (select f ->> 'id' id, f ->> 'type' "type", (f ->> 'delta')::bigint delta, (f ->> 'value')::float "value", f ->> 'hash' hash from jsonb_array_elements(i_list) f),
	b as (select id from metric m where exists (select 0 from a where a.id = m.id for no key update))
	insert into metric(id, "type", delta, "value", hash) select id, "type", delta, "value", hash from a
	on conflict (id) do update set delta = excluded.delta, "value" = excluded."value", hash = excluded.hash;   
end $$;

/*
select metrics_set('[{"id": "1", "type": "gauge", "value": 3.1416}]')

select * from metrics_get('{"id": "o"}'::jsonb)

select * from metric
*/