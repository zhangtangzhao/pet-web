-- 回滚种子数据
DELETE FROM pet_product WHERE id IN (3001, 3002, 3003, 3004, 3005);
DELETE FROM breed WHERE id IN (2011, 2012, 2013, 2021, 2022, 2023, 2031, 2041);
DELETE FROM category WHERE id IN (101, 102, 103, 104);
DELETE FROM admin_user WHERE username = 'admin';
