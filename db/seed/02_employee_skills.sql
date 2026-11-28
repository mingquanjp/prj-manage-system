-- Hồ sơ năng lực chính thức của nhân sự mẫu.
--
-- Dữ liệu mẫu dùng tài khoản admin@company.vn làm người xác nhận tường minh;
-- đây không phải lịch sử đề nghị hoặc bằng chứng phê duyệt của người dùng thật.

INSERT INTO employee_skills (
    employee_id, skill_id, proficiency_level, years_experience,
    approved_by, approved_at
)
SELECT
    e.employee_id,
    s.skill_id,
    v.level::proficiency_level,
    v.years,
    admin.user_id,
    now()
FROM (VALUES
    -- Backend Developer — Senior
    ('dev.an@company.vn', 'Java',                    'EXPERT',       6.0),
    ('dev.an@company.vn', 'Spring Boot',             'EXPERT',       5.5),
    ('dev.an@company.vn', 'SQL',                     'ADVANCED',     6.0),
    ('dev.an@company.vn', 'PostgreSQL',              'ADVANCED',     4.0),
    ('dev.an@company.vn', 'Thiết kế REST API',       'EXPERT',       5.0),
    ('dev.an@company.vn', 'Kiến trúc Microservices', 'ADVANCED',     3.0),
    ('dev.an@company.vn', 'Docker',                  'INTERMEDIATE', 2.5),

    -- Backend Developer — Middle
    ('dev.binh@company.vn', 'Go',                      'ADVANCED',     3.5),
    ('dev.binh@company.vn', 'PostgreSQL',              'ADVANCED',     3.0),
    ('dev.binh@company.vn', 'SQL',                     'ADVANCED',     3.5),
    ('dev.binh@company.vn', 'Thiết kế REST API',       'ADVANCED',     3.0),
    ('dev.binh@company.vn', 'Redis',                   'INTERMEDIATE', 1.5),
    ('dev.binh@company.vn', 'Docker',                  'INTERMEDIATE', 2.0),

    -- Backend Developer — Junior
    ('dev.cuong@company.vn', 'Node.js',                 'INTERMEDIATE', 1.5),
    ('dev.cuong@company.vn', 'NestJS',                  'BASIC',        1.0),
    ('dev.cuong@company.vn', 'SQL',                     'INTERMEDIATE', 1.5),
    ('dev.cuong@company.vn', 'MySQL',                   'INTERMEDIATE', 1.0),
    ('dev.cuong@company.vn', 'Làm việc nhóm',           'ADVANCED',     1.5),

    -- Frontend Developer — Senior
    ('dev.dung@company.vn', 'JavaScript',              'EXPERT',       6.0),
    ('dev.dung@company.vn', 'TypeScript',              'EXPERT',       4.5),
    ('dev.dung@company.vn', 'React',                   'EXPERT',       5.0),
    ('dev.dung@company.vn', 'Next.js',                 'ADVANCED',     3.0),
    ('dev.dung@company.vn', 'HTML/CSS',                'EXPERT',       6.0),
    ('dev.dung@company.vn', 'Tailwind CSS',            'ADVANCED',     2.5),

    -- Frontend Developer — Middle
    ('dev.giang@company.vn', 'JavaScript',              'ADVANCED',     3.0),
    ('dev.giang@company.vn', 'TypeScript',              'INTERMEDIATE', 2.0),
    ('dev.giang@company.vn', 'React',                   'ADVANCED',     2.5),
    ('dev.giang@company.vn', 'HTML/CSS',                'ADVANCED',     3.0),
    ('dev.giang@company.vn', 'Tailwind CSS',            'INTERMEDIATE', 1.5),

    -- Frontend Developer — Junior
    ('dev.hai@company.vn', 'JavaScript',              'INTERMEDIATE', 1.0),
    ('dev.hai@company.vn', 'HTML/CSS',                'ADVANCED',     1.5),
    ('dev.hai@company.vn', 'Vue.js',                  'INTERMEDIATE', 1.0),
    ('dev.hai@company.vn', 'React',                   'BASIC',        0.5),

    -- Fullstack — Senior
    ('dev.khanh@company.vn', 'TypeScript',              'EXPERT',       5.0),
    ('dev.khanh@company.vn', 'React',                   'ADVANCED',     4.0),
    ('dev.khanh@company.vn', 'Node.js',                 'EXPERT',       5.0),
    ('dev.khanh@company.vn', 'NestJS',                  'ADVANCED',     3.0),
    ('dev.khanh@company.vn', 'PostgreSQL',              'ADVANCED',     4.0),
    ('dev.khanh@company.vn', 'Thiết kế REST API',       'EXPERT',       4.5),
    ('dev.khanh@company.vn', 'Docker',                  'ADVANCED',     3.0),

    -- Fullstack — Middle
    ('dev.linh@company.vn', 'JavaScript',              'ADVANCED',     3.0),
    ('dev.linh@company.vn', 'React',                   'INTERMEDIATE', 2.0),
    ('dev.linh@company.vn', 'Python',                  'ADVANCED',     3.0),
    ('dev.linh@company.vn', 'Django',                  'INTERMEDIATE', 2.0),
    ('dev.linh@company.vn', 'SQL',                     'INTERMEDIATE', 2.5),

    -- Mobile — Middle
    ('dev.minh@company.vn', 'Flutter',                 'ADVANCED',     3.0),
    ('dev.minh@company.vn', 'Android',                 'INTERMEDIATE', 2.0),
    ('dev.minh@company.vn', 'Thiết kế REST API',       'INTERMEDIATE', 2.0),
    ('dev.minh@company.vn', 'Giải quyết vấn đề',       'ADVANCED',     3.0),

    -- Mobile — Junior
    ('dev.ngan@company.vn', 'React Native',            'INTERMEDIATE', 1.0),
    ('dev.ngan@company.vn', 'JavaScript',              'INTERMEDIATE', 1.5),
    ('dev.ngan@company.vn', 'iOS',                     'BASIC',        0.5),

    -- QA — Senior
    ('qa.phuc@company.vn', 'Kiểm thử thủ công',       'EXPERT',       6.0),
    ('qa.phuc@company.vn', 'Kiểm thử tự động',        'ADVANCED',     4.0),
    ('qa.phuc@company.vn', 'Selenium',                'ADVANCED',     4.0),
    ('qa.phuc@company.vn', 'Kiểm thử API',            'EXPERT',       4.5),
    ('qa.phuc@company.vn', 'Kiểm thử hiệu năng',      'INTERMEDIATE', 2.0),
    ('qa.phuc@company.vn', 'SQL',                     'INTERMEDIATE', 3.0),

    -- QA — Middle
    ('qa.quyen@company.vn', 'Kiểm thử thủ công',       'ADVANCED',     2.5),
    ('qa.quyen@company.vn', 'Kiểm thử API',            'INTERMEDIATE', 2.0),
    ('qa.quyen@company.vn', 'Selenium',                'INTERMEDIATE', 1.5),
    ('qa.quyen@company.vn', 'Làm việc nhóm',           'ADVANCED',     2.5),

    -- DevOps — Senior
    ('ops.son@company.vn', 'Docker',                  'EXPERT',       5.0),
    ('ops.son@company.vn', 'Kubernetes',              'ADVANCED',     3.5),
    ('ops.son@company.vn', 'CI/CD',                   'EXPERT',       5.0),
    ('ops.son@company.vn', 'AWS',                     'ADVANCED',     4.0),
    ('ops.son@company.vn', 'Linux',                   'EXPERT',       6.0),
    ('ops.son@company.vn', 'Nginx',                   'ADVANCED',     4.0),

    -- DevOps — Junior
    ('ops.tuan@company.vn', 'Docker',                  'INTERMEDIATE', 1.5),
    ('ops.tuan@company.vn', 'Linux',                   'INTERMEDIATE', 2.0),
    ('ops.tuan@company.vn', 'CI/CD',                   'BASIC',        1.0),

    -- Business Analyst — Senior
    ('ba.uyen@company.vn', 'Phân tích yêu cầu',       'EXPERT',       6.0),
    ('ba.uyen@company.vn', 'UML',                     'ADVANCED',     5.0),
    ('ba.uyen@company.vn', 'Viết tài liệu đặc tả',    'EXPERT',       6.0),
    ('ba.uyen@company.vn', 'Agile/Scrum',             'ADVANCED',     4.0),
    ('ba.uyen@company.vn', 'Giao tiếp',               'EXPERT',       6.0),
    ('ba.uyen@company.vn', 'SQL',                     'INTERMEDIATE', 2.0),

    -- Business Analyst — Junior
    ('ba.vy@company.vn', 'Phân tích yêu cầu',       'INTERMEDIATE', 1.0),
    ('ba.vy@company.vn', 'UML',                     'INTERMEDIATE', 1.0),
    ('ba.vy@company.vn', 'Viết tài liệu đặc tả',    'INTERMEDIATE', 1.5),
    ('ba.vy@company.vn', 'Giao tiếp',               'ADVANCED',     1.5),

    -- Technical Lead — Senior
    ('lead.xuan@company.vn', 'Java',                    'EXPERT',       8.0),
    ('lead.xuan@company.vn', 'Kiến trúc Microservices', 'EXPERT',       5.0),
    ('lead.xuan@company.vn', 'Thiết kế cơ sở dữ liệu',  'EXPERT',       7.0),
    ('lead.xuan@company.vn', 'PostgreSQL',              'EXPERT',       6.0),
    ('lead.xuan@company.vn', 'Docker',                  'ADVANCED',     4.0),
    ('lead.xuan@company.vn', 'Agile/Scrum',             'ADVANCED',     5.0),
    ('lead.xuan@company.vn', 'Giải quyết vấn đề',       'EXPERT',       8.0),

    -- Hai Quản lý dự án cũng có hồ sơ năng lực để tra cứu
    ('pm.hoa@company.vn', 'Agile/Scrum',             'EXPERT',       6.0),
    ('pm.hoa@company.vn', 'Phân tích yêu cầu',       'ADVANCED',     5.0),
    ('pm.hoa@company.vn', 'Giao tiếp',               'EXPERT',       7.0),
    ('pm.hoa@company.vn', 'Quản lý thời gian',       'EXPERT',       6.0),

    ('pm.nam@company.vn', 'Agile/Scrum',             'ADVANCED',     4.0),
    ('pm.nam@company.vn', 'Java',                    'ADVANCED',     5.0),
    ('pm.nam@company.vn', 'Quản lý thời gian',       'ADVANCED',     4.0),
    ('pm.nam@company.vn', 'Giao tiếp',               'ADVANCED',     5.0)
) AS v(company_email, skill_name, level, years)
JOIN employees e   ON lower(btrim(e.email)) = v.company_email
JOIN skills    s   ON s.skill_name    = v.skill_name
CROSS JOIN LATERAL (SELECT user_id FROM employees WHERE lower(btrim(email)) = 'admin@company.vn') AS admin
ON CONFLICT ON CONSTRAINT uq_employee_skills DO NOTHING;
