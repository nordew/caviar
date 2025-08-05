
BEGIN;

-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    slug VARCHAR(255) NOT NULL,
    name VARCHAR(500) NOT NULL,
    subtitle VARCHAR(1000),
    description TEXT,
    images JSONB DEFAULT '[]'::jsonb,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create product_variants table
CREATE TABLE IF NOT EXISTS product_variants (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL,
    mass INTEGER NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0,
    prices JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_product_variants_product_id 
        FOREIGN KEY (product_id) REFERENCES products(id) 
        ON DELETE CASCADE ON UPDATE CASCADE
);

-- Essential indexes for performance
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_slug 
ON products (slug);

CREATE INDEX IF NOT EXISTS idx_product_variants_product_id 
ON product_variants (product_id);

CREATE INDEX IF NOT EXISTS idx_product_variants_product_mass 
ON product_variants (product_id, mass);

-- Create triggers for updated_at
CREATE TRIGGER trigger_products_updated_at 
    BEFORE UPDATE ON products 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_product_variants_updated_at 
    BEFORE UPDATE ON product_variants 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;