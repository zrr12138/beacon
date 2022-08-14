//
namespace beacon cpp
namespace py cpp

const set<string> Resource_type={'wood','stone','iron','food'}
struct Soldier_info
{
    1: i64 blood,
    2: i64 attack,
    3: i64 defen,
    4: i64 speed,
    5: i64 burden,
    6: i64 camp_required,
    7: map<string,i64> resources_need,
    8: i64 time_need,
    9: i64 waste,
}
/*
    Transport  运输兵
    Knife      短刀兵
    Gun        长抢兵
    Scout      斥候
    Rider      轻骑兵
    Shield     盾弓手
    Heavy      重装步兵
*/
const set<string> Soldier_type = {'transport','knife','gun','scout','rider','shield','heavy'}

struct Coordinate
{
    1: i64 x,
    2: i64 y,
}
//行军
struct Military 
{
    1: i64 u_id,
    2: map<string,i64> soldiers, //兵种-个数
    9: Coordinate aim,
    10: map<string,i64> burden, //资源-数量
    11: i64 time_stamp
}
//招募
struct Recute 
{
    1: i64 u_id,
    2: string soldier_type,
    3: i64 left_num,
    4: i64 time_stamp
}
struct Wait_recute
{
    1: i64 u_id,
    2: i64 id,//等待顺序
    3: string soldier_type,
    4: i64 left_num,
}
const set<string> Building_type = {'government','camp','storehouse','granary','defence',
                                    'wood_house','stone_house','iron_house','food_house'}
struct City_build_grade
{
    1: i64 u_id,
    2:map<string,i64> buildings,//建筑-等级
}
struct City_resources_num
{
    1:i64 u_id,
    2:map<string,i64> num, //资源-数量
    3:map<string,i64> max_num, //资源-上限
}
struct City_Military 
{
    1:i64 u_id,
    2: map<string,i64> soldiers, //兵种-个数
}
//升级建筑
struct Update_building
{
    1: i64 u_id,
    2: string building_type,
    3: i64 time_stamp
}
struct Wait_update_building
{
    1: i64 u_id,
    2: i64 id,//等待顺序
    3: string building_type,
}
const set<string> Location_type={'city','waste'}
struct Location
{
    1: Coordinate coordinate,
    2: string location_type,
    3: i64 u_id, //type=city，该字段有意义，表示所属用户
    4: i64 grade,//type用于扩展
}
