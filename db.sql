-- 書籍基本情報
CREATE TABLE books (
  id VARCHAR(36) PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  author VARCHAR(255),
  publisher VARCHAR(255),
  isbn VARCHAR(20),
  publication_date DATE,
  image_url VARCHAR(255),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- ECサイト情報
CREATE TABLE sites (
  id VARCHAR(36) PRIMARY KEY,
  name VARCHAR(50) NOT NULL,
  base_url VARCHAR(255) NOT NULL,
  affiliate_id VARCHAR(100)
);

-- 書籍のECサイト個別情報
CREATE TABLE book_site_mappings (
  id VARCHAR(36) PRIMARY KEY,
  book_id VARCHAR(36) NOT NULL,
  site_id VARCHAR(36) NOT NULL,
  site_specific_id VARCHAR(100) NOT NULL,
  price DECIMAL(10, 2),
  url VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (book_id) REFERENCES books(id),
  FOREIGN KEY (site_id) REFERENCES sites(id)
);

-- カテゴリ情報
CREATE TABLE categories (
  id VARCHAR(36) PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  parent_id VARCHAR(36),
  FOREIGN KEY (parent_id) REFERENCES categories(id)
);

-- サイト別カテゴリマッピング
CREATE TABLE site_category_mappings (
  id VARCHAR(36) PRIMARY KEY,
  category_id VARCHAR(36) NOT NULL,
  site_id VARCHAR(36) NOT NULL,
  site_specific_category_id VARCHAR(100) NOT NULL,
  FOREIGN KEY (category_id) REFERENCES categories(id),
  FOREIGN KEY (site_id) REFERENCES sites(id)
);

-- ランキングデータ
CREATE TABLE rankings (
  id VARCHAR(36) PRIMARY KEY,
  book_site_mapping_id VARCHAR(36) NOT NULL,
  category_id VARCHAR(36) NOT NULL,
  rank INT NOT NULL,
  period_type ENUM('daily', 'weekly', 'monthly', 'yearly') NOT NULL,
  date_from DATE NOT NULL,
  date_to DATE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (book_site_mapping_id) REFERENCES book_site_mappings(id),
  FOREIGN KEY (category_id) REFERENCES categories(id),
  INDEX idx_rank_period (rank, period_type, date_from),
  INDEX idx_category_period (category_id, period_type, date_from)
);