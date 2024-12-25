create table packages
(
    name text not null,
    license text not null,
    version text not null,
    distributor text not null,
    distribution_url text not null,
    constraint packages_pk primary key (name, version, distributor)
);
