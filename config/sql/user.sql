CREATE TABLE users (
    user_id BIGINT PRIMARY KEY,          -- 用户ID，主键，使用雪花算法生成
    email VARCHAR(255) NOT NULL UNIQUE,  -- 用户邮箱，唯一
    password_hash VARCHAR(255) NOT NULL, -- 用户密码的哈希值
    username VARCHAR(50) NOT NULL,       -- 用户名
    first_name VARCHAR(50),              -- 用户的名
    last_name VARCHAR(50),               -- 用户的姓
    phone_number VARCHAR(20),            -- 用户的电话号码
    address TEXT,                        -- 用户的地址
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP -- 更新时间
);
