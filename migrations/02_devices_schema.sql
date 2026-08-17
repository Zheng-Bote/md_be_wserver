-- SPDX-FileComment: Medical Devices Backend
-- SPDX-FileType: SOURCE
-- SPDX-FileContributor: ZHENG Robert
-- SPDX-FileCopyrightText: 2026 ZHENG Robert
-- SPDX-License-Identifier: Apache-2.0

-- Create Device Types Table
CREATE TABLE IF NOT EXISTS device_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_update TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN NOT NULL DEFAULT true
);

-- Create Devices Table
CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type_id UUID REFERENCES device_types(id) ON DELETE SET NULL,
    device_name VARCHAR(255) NOT NULL,
    manufacturer VARCHAR(255),
    interface VARCHAR(50),
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_update TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN NOT NULL DEFAULT true
);

-- Setup trigger to automatically update 'last_update'
CREATE OR REPLACE FUNCTION update_last_update_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_update = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_device_types_last_update ON device_types;
CREATE TRIGGER update_device_types_last_update
    BEFORE UPDATE ON device_types
    FOR EACH ROW
    EXECUTE FUNCTION update_last_update_column();

DROP TRIGGER IF EXISTS update_devices_last_update ON devices;
CREATE TRIGGER update_devices_last_update
    BEFORE UPDATE ON devices
    FOR EACH ROW
    EXECUTE FUNCTION update_last_update_column();

-- Seed Data for Device Types
INSERT INTO device_types (name, description) VALUES
('Imaging modalities', 'Computed tomography (CT), magnetic resonance imaging (MRI), sonography (ultrasound), and X-ray diagnostics'),
('Functional and vital sign monitoring', 'ECG, EEG, blood pressure monitors, and pulse oximeters'),
('Laboratory diagnostics & in vitro devices', 'Blood gas analysis, analyzers for clinical chemistry or pathology'),
('Endoscopy', 'Optical systems for the direct visualization of body cavities and organs')
ON CONFLICT (name) DO NOTHING;

-- Seed Data for Devices (Unique Entries)
INSERT INTO devices (device_name, type_id, manufacturer, interface, description) VALUES
('Maico MA33', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'DIATEC AG', 'GDT', 'Online connection of hearing test devices (air and bone conduction). Transfer of measurement results into the subject file incl. curve display.'),
('Custo Cardio 400', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'custo med GmbH', 'GDT', 'Import of resting ECG data from various manufacturers, structured storage in the medical documentation.'),
('Custo ec 5000', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'custo med GmbH', 'GDT', 'Transmission of stress and performance data, frequently coupled with ECG systems.'),
('Custo Spiro mobile', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'custo med GmbH', 'GDT', 'Import of pulmonary function measurements incl. curves and reference values.'),
('Custo Cardio 400', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'custo med GmbH', 'GDT', 'Transfer of long-term ECG analysis and summary reports.'),
('Custo Cardio 400', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'custo med GmbH', 'GDT', 'Transmission of 24h blood pressure measurements incl. mean value and time series analysis.'),
('Perivist FeV,Perivist compact II', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'Vistec GmbH', 'GDT', 'Import of perimetry measurement data for assessing the visual field.'),
('Optovist EU,OPTOVIST II', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'Vistec GmbH', 'GDT', 'Transmission of visual acuity, contrast and, if applicable, color vision values.'),
('Samsung HS 40, Mindray Z60', (SELECT id FROM device_types WHERE name = 'Imaging modalities'), 'Samsung,Mindray Medical Germany GmbH', 'DICOM', 'A DICOM interface is simulated via software (sonoGDT). With this software, employee data from the occupational health software can be transmitted to the ultrasound device via the GDT interface. Once the examination is complete, images and video clips can be transmitted to the occupational health software.'),
('MiniScreenPlus,Samoa', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'Löwenstein Medical', 'GDT', 'Import of polygraphic measurement values and evaluations.'),
('i Lab Client', (SELECT id FROM device_types WHERE name = 'Laboratory diagnostics & in vitro devices'), 'SynLab, itech Laborlösungen GmbH', 'LDT', 'Import of incoming laboratory values with assignment to the subject file.'),
('CA850', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'Tre Metrics', 'Y', 'Amplivox'),
('MSA AirHawk System', NULL, 'MSA', 'N/A', 'SCBA'),
('AS216', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'NS Clinical Technologies', 'N', 'an instrument used to measure hearing ability'),
('SA-232X', (SELECT id FROM device_types WHERE name = 'Laboratory diagnostics & in vitro devices'), 'Sturdy Industrial', 'N', 'An autoclave is a machine that uses steam under pressure to kill harmful bacteria, viruses, fungi, and spores'),
('KOKO PFT', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'SSEM Mthembu Medical', 'N', 'An instrument used to measure the air capacity in the lungs'),
('VS-V', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'NS Clinical Technologies', 'N', 'a type of vision screening tool used to assess various aspects of an individual''s vision, such as acuity, binocular vision, and peripheral vision'),
('TEC-5521K', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'Nihon Kohden', 'N', 'a medical device used to restore a normal heart rhythm by delivering an electric shock to the heart'),
('MDW-250L', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'Adam', 'N', 'An instrument used to measure body weight'),
('MDW-300L', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'Adam', 'N', 'An instrument used to measure body weight'),
('SC0020', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'MediSensor', 'N', 'An instrument used to measure body weight'),
('Northern Aquarius', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'Northern Aquarius', 'N', 'A device used to measure blood pressure'),
('M4', (SELECT id FROM device_types WHERE name = 'Functional and vital sign monitoring'), 'Omron', 'N', 'A device used to measure blood pressure');
