-- 生成商品分类数据
INSERT INTO product_categories (id, name, description, sort_order, created_at, updated_at) VALUES
(1, '手机', '各类智能手机及配件', 10, NOW(), NOW()),
(2, '电脑', '笔记本电脑、台式机及配件', 20, NOW(), NOW()),
(3, '家电', '家用电器及智能家居产品', 30, NOW(), NOW()),
(4, '服装', '男女服装、鞋帽及配饰', 40, NOW(), NOW()),
(5, '美妆', '化妆品、护肤品及个人护理', 50, NOW(), NOW()),
(6, '食品', '零食、饮料及生鲜食品', 60, NOW(), NOW());

-- 生成商品数据
INSERT INTO products (id, name, description, price, stock, category_id, is_published, published_at, images, created_at, updated_at) VALUES
(1, 'iPhone 15 Pro Max', '苹果最新旗舰手机，搭载A17芯片，超强性能，出色的拍照体验。', 8999.00, 100, 1, true, NOW(), '["images/products/iphone15pm_1.jpg", "images/products/iphone15pm_2.jpg"]', NOW(), NOW()),
(2, 'MacBook Pro 16英寸', 'Apple M3 Pro芯片，16GB统一内存，1TB固态硬盘，专业级性能。', 18999.00, 50, 2, true, NOW(), '["images/products/macbookpro_1.jpg", "images/products/macbookpro_2.jpg"]', NOW(), NOW()),
(3, '华为Mate 60 Pro', '搭载麒麟芯片，超长续航，专业影像系统，卫星通信。', 6999.00, 80, 1, true, NOW(), '["images/products/huaweimate60_1.jpg", "images/products/huaweimate60_2.jpg"]', NOW(), NOW()),
(4, '小米电视大师 77英寸OLED', '4K超高清OLED屏幕，120Hz高刷新率，杜比视界，智能语音控制。', 19999.00, 30, 3, true, NOW(), '["images/products/mitv_1.jpg", "images/products/mitv_2.jpg"]', NOW(), NOW()),
(5, 'NIKE Air Jordan 1', '经典高帮篮球鞋，舒适耐穿，时尚百搭。', 1299.00, 200, 4, true, NOW(), '["images/products/nike_aj1_1.jpg", "images/products/nike_aj1_2.jpg"]', NOW(), NOW()),
(6, '兰蔻小黑瓶精华 50ml', '高效抗老精华，快速修护，提亮肤色，改善肤质。', 899.00, 150, 5, true, NOW(), '["images/products/lancome_1.jpg", "images/products/lancome_2.jpg"]', NOW(), NOW()),
(7, '三只松鼠坚果大礼包', '多种坚果组合，营养美味，送礼佳选。', 149.00, 300, 6, true, NOW(), '["images/products/threesquirrels_1.jpg", "images/products/threesquirrels_2.jpg"]', NOW(), NOW()),
(8, '戴尔XPS 15笔记本', '英特尔i9处理器，32GB内存，1TB SSD，4K触控屏，专业创作利器。', 15999.00, 40, 2, true, NOW(), '["images/products/dellxps_1.jpg", "images/products/dellxps_2.jpg"]', NOW(), NOW()),
(9, '海尔变频冰箱', '风冷无霜，多维智能控温，大容量，节能环保。', 3999.00, 60, 3, true, NOW(), '["images/products/haier_1.jpg", "images/products/haier_2.jpg"]', NOW(), NOW()),
(10, '优衣库男士休闲裤', '舒适面料，简约设计，百搭款式，多色可选。', 199.00, 500, 4, true, NOW(), '["images/products/uniqlo_1.jpg", "images/products/uniqlo_2.jpg"]', NOW(), NOW()),
(11, '红米Note 12 Pro', '高性价比智能手机，1亿像素相机，5000mAh大电池。', 1699.00, 200, 1, true, NOW(), '["images/products/redmi_1.jpg", "images/products/redmi_2.jpg"]', NOW(), NOW()),
(12, 'iPad Air 5', '全面屏设计，M1芯片，轻薄便携，多用途平板电脑。', 4299.00, 80, 2, true, NOW(), '["images/products/ipadair_1.jpg", "images/products/ipadair_2.jpg"]', NOW(), NOW()),
(13, '美的空调', '变频节能，智能控制，强效制冷/热，静音运行。', 2999.00, 70, 3, true, NOW(), '["images/products/midea_1.jpg", "images/products/midea_2.jpg"]', NOW(), NOW()),
(14, 'SK-II神仙水 230ml', '明星产品，提亮肤色，改善肤质，提升肌肤透明度。', 1599.00, 120, 5, true, NOW(), '["images/products/skii_1.jpg", "images/products/skii_2.jpg"]', NOW(), NOW()),
(15, '良品铺子零食大礼包', '多种休闲零食组合，味道丰富，送礼自用两相宜。', 99.00, 400, 6, true, NOW(), '["images/products/liangpin_1.jpg", "images/products/liangpin_2.jpg"]', NOW(), NOW());

-- 生成商品SKU数据
INSERT INTO product_skus (id, product_id, sku, price, stock, specs, created_at, updated_at) VALUES
(1, 1, 'IP15PM-256G-BLACK', 8999.00, 30, '{"color": "黑色", "storage": "256GB"}', NOW(), NOW()),
(2, 1, 'IP15PM-256G-SILVER', 8999.00, 30, '{"color": "银色", "storage": "256GB"}', NOW(), NOW()),
(3, 1, 'IP15PM-512G-BLACK', 9999.00, 20, '{"color": "黑色", "storage": "512GB"}', NOW(), NOW()),
(4, 1, 'IP15PM-512G-SILVER', 9999.00, 20, '{"color": "银色", "storage": "512GB"}', NOW(), NOW()),
(5, 2, 'MBP16-M3P-16G-1T-SPACE', 18999.00, 25, '{"color": "深空灰", "memory": "16GB", "storage": "1TB"}', NOW(), NOW()),
(6, 2, 'MBP16-M3P-16G-1T-SILVER', 18999.00, 25, '{"color": "银色", "memory": "16GB", "storage": "1TB"}', NOW(), NOW()),
(7, 3, 'HW-M60P-256G-BLACK', 6999.00, 40, '{"color": "黑色", "storage": "256GB"}', NOW(), NOW()),
(8, 3, 'HW-M60P-512G-BLACK', 7699.00, 40, '{"color": "黑色", "storage": "512GB"}', NOW(), NOW()),
(9, 5, 'NK-AJ1-40-RED', 1299.00, 50, '{"color": "红色", "size": "40"}', NOW(), NOW()),
(10, 5, 'NK-AJ1-41-RED', 1299.00, 50, '{"color": "红色", "size": "41"}', NOW(), NOW()),
(11, 5, 'NK-AJ1-42-RED', 1299.00, 50, '{"color": "红色", "size": "42"}', NOW(), NOW()),
(12, 5, 'NK-AJ1-40-BLACK', 1299.00, 50, '{"color": "黑色", "size": "40"}', NOW(), NOW()),
(13, 6, 'LC-XHF-50ML', 899.00, 150, '{"size": "50ml"}', NOW(), NOW()),
(14, 11, 'XM-N12P-128G-BLUE', 1699.00, 100, '{"color": "蓝色", "storage": "128GB"}', NOW(), NOW()),
(15, 11, 'XM-N12P-128G-BLACK', 1699.00, 100, '{"color": "黑色", "storage": "128GB"}', NOW(), NOW()),
(16, 12, 'IPAD-AIR5-64G-GRAY', 4299.00, 40, '{"color": "深空灰", "storage": "64GB", "connectivity": "WiFi"}', NOW(), NOW()),
(17, 12, 'IPAD-AIR5-64G-BLUE', 4299.00, 40, '{"color": "蓝色", "storage": "64GB", "connectivity": "WiFi"}', NOW(), NOW()); 