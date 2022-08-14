create table Military 
(
    u_id bigint primary key,
    transport bigint,
    knife bigint,
    gun bigint,
    scout bigint,
    rider bigint,
    shield bigint,
    heavy bigint,
    x bigint,
    y bigint,
    wood bigint,
    stone bigint,
    iron bigint,
    food bigint,
    time_stamp bigint
);
create table Recute
(
    u_id bigint primary key,
    soldier_type varchar(40),
    left_num bigint,
    time_stamp bigint,
);
create table Wait_recute
(
    u_id bigint primary key,
    id bigint,
    soldier_type varchar(40),
    left_num bigint,
);
create table City_build_grade
(
    u_id bigint primary key,
    government bigint,
    camp bigint,
    storehouse bigint,
    granary bigint,
    defence bigint,
    wood_house bigint,
    stone_house bigint,
    iron_house bigint,
    food_house bigint,
);
create table City_resources_num
(
    u_id bigint primary key,
    wood bigint,
    stone bigint,
    iron bigint,
    food bigint,
    wood_max bigint,
    stone_max bigint,
    iron_max bigint,
    food_max bigint,
);
create table City_Military 
(
    u_id bigint primary key,
    transport bigint,
    knife bigint,
    gun bigint,
    scout bigint,
    rider bigint,
    shield bigint,
    heavy bigint,
);
create table Update_building
(
    u_id bigint primary key,
    building_type varchar(40),
    time_stamp bigint,
);
create table Wait_update_building
(
    u_id bigint primary key,
    id bigint,
    building_type varchar(40),
);
create table Location_
(
    x bigint,
    y bigint,
    location_type varchar(40),
    u_id bigint,
    grade bigint,
);
