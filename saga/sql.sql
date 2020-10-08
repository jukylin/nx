CREATE TABLE txcompensate (
  id int COLLATE utf8mb4_general_ci not NULL,
  txid unsigned bigint Not NULL default 0 COMMENT '事务编号',
  success int not NULL default 0,
  step int not NULL,
  create_time datetime not NULL DEFAULT CURRENT_TIMESTAMP,
  update_time datetime not NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted TINYINT(1) UNSIGNED NOT NULL DEFAULT '0' COMMENT '删除标识',
  PRIMARY KEY (id) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 comment="补偿结果" COLLATE=utf8mb4_general_ci;
CREATE TABLE txgroup (
  id int COLLATE utf8mb4_general_ci not NULL,
  txid unsigned bigint Not NULL default 0 COMMENT '事务编号',
  state int not NULL,
  priority int not NULL,
  create_time datetime not NULL,
  update_time datetime not NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted TINYINT(1) UNSIGNED NOT NULL DEFAULT '0' COMMENT '删除标识',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 comment="事物主表" COLLATE=utf8mb4_general_ci;
CREATE TABLE txrecord (
  id int COLLATE utf8mb4_general_ci not NULL,
  txid unsigned bigint Not NULL default 0 COMMENT '事务编号',
  manner_name varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci not NULL,
  method_name varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci not NULL,
  compensate_name varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci not NULL,
  class_name varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci not NULL,
  service_name varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci not NULL,
  generic_param_types varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci not NULL,
  param_types varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci not NULL,
  params varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci not NULL,
  step smallint not NULL,
  lookup varchar(255) COLLATE utf8mb4_general_ci not NULL,
  reg_address varchar(500) COLLATE utf8mb4_general_ci not NULL,
  version varchar(255) COLLATE utf8mb4_general_ci not NULL,
  create_time datetime not NULL DEFAULT CURRENT_TIMESTAMP,
  update_time datetime not NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted TINYINT(1) UNSIGNED NOT NULL DEFAULT '0' COMMENT '删除标识',
  PRIMARY KEY (id) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 comment="事物步骤" COLLATE=utf8mb4_general_ci;
