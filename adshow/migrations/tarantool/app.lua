box.cfg{listen = 3301}
box.schema.user.passwd('pass')
my_space = box.schema.create_space("platforms", { if_not_exists = true })
my_space:format({{ name = 'platform_id', type = 'number' }, { name = 'device_type', type = 'string'}, { name = 'banner_id', type = 'number'}})

my_space:create_index('primary', {type = 'tree', parts = {'platform_id', 'device_type', 'banner_id'}, if_not_exists = true })
my_space:create_index('secondary', {type = 'tree', parts = {'banner_id'}, unique = false, if_not_exists = true })