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
	with a as (select f ->> 'id' id, f ->> 'type' "type", (f ->> 'delta')::bigint delta, (f ->> 'value')::float "value", f ->> 'hash' hash, i from jsonb_array_elements(i_list) with ordinality f(f, i)),
	aa as (
		select id, min("type") "type", sum(delta) delta, null::float "value", null::text hash from a where "type" = 'counter' group by id
	),
	aaa as (		
		select distinct on (id) id, "type", null::bigint delta, "value", hash from a where "type" = 'gauge' order by id, i desc
	),
	aaaa as (select * from aa union select * from aaa),
	b as (select id from metric m where exists (select 0 from aaaa where aaaa.id = m.id for no key update))
	insert into metric(id, "type", delta, "value", hash) select id, "type", delta, "value", hash from aaaa
	on conflict (id) do update set 
		delta = case when excluded."type" = 'counter' then coalesce(metric.delta, 0) + excluded.delta else excluded.delta end, 
		"value" = excluded."value", 
		hash = excluded.hash;   
end $$;