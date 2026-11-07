import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "scripts"))

from render_sql_exporter_config import render_configuration


class FleetInventoryRendererTests(unittest.TestCase):
    def setUp(self):
        self.temporary_directory = tempfile.TemporaryDirectory()
        self.root = Path(self.temporary_directory.name)
        self.inventory = self.root / "fleet.csv"
        self.config = self.root / "config.yml"
        self.config.write_text("global: {}\ncollectors:\n  - collector_name: sample\n", encoding="utf-8")

    def tearDown(self):
        self.temporary_directory.cleanup()

    def test_renders_sql_targets_without_bom_and_encodes_credentials(self):
        self.inventory.write_text(
            "# inventory comment\n"
            "engine,instance_name,host_or_ip,port,environment,cluster_name,username,password,database_name\n"
            "sqlserver,sql-node,db.example,5000,production,cluster-a,monitor,p@ss$:/,master\n"
            "postgres,pg-node,pg.example,5432,production,cluster-b,monitor,unused,postgres\n",
            encoding="utf-8",
        )

        count = render_configuration(self.inventory, self.config)
        output = self.config.read_bytes()

        self.assertEqual(count, 1)
        self.assertFalse(output.startswith(b"\xef\xbb\xbf"))
        self.assertIn(b"sql-node", output)
        self.assertIn(b"p%40ss%24%3A%2F", output)
        self.assertNotIn(b"pg-node", output)
        self.assertIn(b"  - collector_name: sample", output)

    def test_rejects_duplicate_sql_server_instance_names(self):
        header = "engine,instance_name,host_or_ip,port,environment,cluster_name,username,password,database_name\n"
        self.inventory.write_text(
            header
            + "sqlserver,node,db1,1433,production,cluster,monitor,secret,master\n"
            + "sqlserver,node,db2,1433,production,cluster,monitor,secret,master\n",
            encoding="utf-8",
        )

        with self.assertRaisesRegex(ValueError, "Duplicate SQL Server instance_name"):
            render_configuration(self.inventory, self.config)

    def test_rejects_example_credential_placeholders(self):
        self.inventory.write_text(
            "engine,instance_name,host_or_ip,port,environment,cluster_name,username,password,database_name\n"
            "sqlserver,node,db.example,1433,development,cluster,REPLACE_WITH_USER,REPLACE_WITH_SECRET,master\n",
            encoding="utf-8",
        )

        with self.assertRaisesRegex(ValueError, "example placeholders: password, username"):
            render_configuration(self.inventory, self.config)

    def test_uses_example_template_when_generated_config_is_absent(self):
        self.inventory.write_text(
            "engine,instance_name,host_or_ip,port,environment,cluster_name,username,password,database_name\n"
            "sqlserver,node,db.example,1433,development,cluster,monitor,local-secret,master\n",
            encoding="utf-8",
        )
        self.config.unlink()
        example = self.config.with_name("config.example.yml")
        example.write_text("collectors:\n  - collector_name: sample\n", encoding="utf-8")

        count = render_configuration(self.inventory, self.config)

        self.assertEqual(count, 1)
        self.assertTrue(self.config.is_file())
        self.assertIn("collector_name: sample", self.config.read_text(encoding="utf-8"))

    def test_does_not_overwrite_config_when_inventory_has_no_sql_server(self):
        self.inventory.write_text(
            "engine,instance_name,host_or_ip,port,environment,cluster_name,username,password,database_name\n"
            "postgres,pg-node,pg.example,5432,production,cluster,monitor,unused,postgres\n",
            encoding="utf-8",
        )
        original_config = self.config.read_bytes()

        with self.assertRaisesRegex(ValueError, "No SQL Server instances found"):
            render_configuration(self.inventory, self.config)

        self.assertEqual(self.config.read_bytes(), original_config)


if __name__ == "__main__":
    unittest.main()